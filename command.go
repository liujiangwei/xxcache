package xxcache

import (
	"errors"
	"github.com/liujiangwei/xxcache/redis"
)

type Command interface {
	Serialize() redis.Message

	ParseResponse(message redis.Message)

	WithError(err error)
}

type BaseCommand struct {
	args  []string
	err   error
	reply redis.Message
}

func (cmd BaseCommand) Serialize() redis.Message {
	return redis.ConvertToMessage(cmd.args...)
}

func (cmd BaseCommand) WithError(err error) {
	cmd.err = err
}

func NewBaseCommand(args ...string) BaseCommand {
	return BaseCommand{
		args:  args,
		err:   nil,
		reply: nil,
	}
}

type StringCommand struct {
	BaseCommand
	val string
}

func (cmd *StringCommand) ParseResponse(message redis.Message) {
	cmd.val = message.String()
}

func NewStringCommand(args ...string) StringCommand {
	return StringCommand{
		BaseCommand: NewBaseCommand(args...),
	}
}

// for command return ArrayMessage
type StringStringCommand struct {
	BaseCommand
	val map[string]string
}

func (cmd *StringStringCommand) ParseResponse(message redis.Message) {
	messages, ok := message.(redis.ArrayMessage)
	if !ok {
		cmd.err = errors.New("failed to parse response")
		return
	}

	if len(messages)%2 != 0 {
		cmd.err = errors.New("failed to parse response")
		return
	}

	cmd.val = make(map[string]string, len(messages)/2)
	for id := 0; id < len(messages); id += 2 {
		cmd.val[messages[id].String()] = messages[id+1].String()
	}
}

// map[string]string
func NewStringStringCommand(args ...string) StringStringCommand {
	return StringStringCommand{
		BaseCommand: NewBaseCommand(args...),
		val:         make(map[string]string),
	}
}

// for command return -OK
type OKCommand struct {
	BaseCommand
	ok  bool
	val string
}

func (cmd *OKCommand) ParserResponse(message redis.Message) {
	m, ok := message.(redis.SimpleStringMessage)
	if !ok {
		cmd.err = errors.New("failed parse redis data")
	}

	if m.String() != "OK" {
		cmd.err = errors.New(m.String())
	}

	cmd.ok = true
}
