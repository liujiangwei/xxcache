package command

import (
	"github.com/liujiangwei/xxcache/protocol"
	"strings"
)

func Convert(message protocol.Message) RedisCommand {
	var args protocol.ArrayMessage

	if message, ok := message.(protocol.ArrayMessage); ok{
		args = message
	}else{
		args = protocol.ArrayMessage{Data:[]protocol.Message{message}}
	}

	command := RedisCommand{
		args: &args,
	}

	if handler, ok := handlerMap[strings.ToUpper(args.Data[0].String())]; ok{
		command.handler = handler
	}else{
		command.handler = notFound
	}

	return command
}

var handlerMap = map[string]Handler{
	"PING": Ping,
	"SET": Set,
	"GET": Get,
}