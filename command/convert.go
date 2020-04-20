package command

import (
	"github.com/liujiangwei/xxcache/rconn"
	"strings"
)

func Convert(message rconn.Message) RedisCommand {
	var args rconn.ArrayMessage

	if message, ok := message.(rconn.ArrayMessage); ok{
		args = message
	}else{
		args = []rconn.Message{message}
	}

	command := RedisCommand{
		args: args,
	}

	if handler, ok := handlerMap[strings.ToUpper(args[0].String())]; ok{
		command.handler = handler
	}else{
		command.handler = notFound
	}

	return command
}

func ConvertToMessage(args ...string) rconn.Message{
	msg := rconn.ArrayMessage{}
	for _, arg := range args{
		msg  = append(msg, rconn.NewBulkStringMessage(arg))
	}
	return msg
}

var handlerMap = map[string]Handler{
	"PING": Ping,
	"SET": Set,
	"GET": Get,
}