package redis


import (
	"errors"
)

type Command interface {
	// convert message to redis command
	Serialize() Message
	// parse response
	Parse(message Message)
	Error(err error)
	//ReadOnly() bool
	//Read(entry Entry)
	//Set() bool
}


type BaseCommand struct {
	Args []string
	Err  error
}

func (cmd *BaseCommand) Serialize() Message {
	return ConvertToMessage(cmd.Args...)
}

func (cmd *BaseCommand) Error(err error) {
	cmd.Err = err
}

func NewBaseCommand(args ...string) BaseCommand {
	return BaseCommand{
		Args: args,
	}
}

type StringCommand struct {
	BaseCommand
	Val string
}

func (cmd *StringCommand) Parse(message Message) {
	cmd.Val = message.String()
}

// for command return ArrayMessage
type StringStringCommand struct {
	BaseCommand
	val map[string]string
}

func (cmd *StringStringCommand) Parse(message Message) {
	messages, ok := message.(ArrayMessage)
	if !ok {
		cmd.Err = errors.New("failed to parse response")
		return
	}

	if len(messages)%2 != 0 {
		cmd.Err = errors.New("failed to parse response")
		return
	}

	cmd.val = make(map[string]string, len(messages)/2)
	for id := 0; id < len(messages); id += 2 {
		cmd.val[messages[id].String()] = messages[id+1].String()
	}
}


// for command return -OK
type OKCommand struct {
	BaseCommand
	val string
}

func (cmd *OKCommand) Parse(message Message) {
	m, ok := message.(SimpleStringMessage)
	if !ok {
		cmd.Err = errors.New("failed parse redis data")
	}

	if m.String() != "OK" {
		cmd.Err = errors.New(m.String())
	}
}

func (cmd *OKCommand) Result() (string, error){
	return cmd.val, cmd.Err
}
