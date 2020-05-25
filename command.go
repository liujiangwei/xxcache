package xxcache

import (
	"errors"
	"github.com/liujiangwei/xxcache/redis"
)

type Command interface {
	Serialize() redis.Message

	ParseResponse(message redis.Message) error
}

type BaseCommand struct {
	args  []string
	err   error
	reply redis.Message
}

func (cmd BaseCommand) Serialize() redis.Message {
	return redis.ConvertToMessage(cmd.args...)
}

func NewBaseCommand(args ...string) BaseCommand{
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

func (cmd *StringCommand) ParseResponse(message redis.Message) error{
	cmd.val = message.String()

	return nil
}

func (cmd *StringCommand) Result() string{
	return cmd.val
}

func NewStringCommand(args ...string) StringCommand{
	return StringCommand{
		BaseCommand: NewBaseCommand(args...),
	}
}


// string string map
type StringStringCommand struct {
	BaseCommand
	val map[string]string
}

func (cmd *StringStringCommand) ParseResponse(message redis.Message) error{
	messages, ok := message.(redis.ArrayMessage)
	if !ok{
		return errors.New("failed to parse response")
	}

	if len(messages) % 2 != 0{
		return errors.New("failed to parse response")
	}

	cmd.val = make(map[string]string, len(messages) / 2)
	for id := 0; id < len(messages);  id += 2{
		cmd.val[messages[id].String()] = messages[id + 1].String()
	}

	return nil
}

func (cmd *StringStringCommand)Result() map[string]string{
	return cmd.val
}

// map[string]string
func NewStringStringCommand(args ...string) StringStringCommand{
	return StringStringCommand{
		BaseCommand: NewBaseCommand(args...),
		val:make(map[string]string),
	}
}

