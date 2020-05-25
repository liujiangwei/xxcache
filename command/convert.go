package command

import (
	"github.com/liujiangwei/xxcache/redis"
	"strings"
)

func Convert(message redis.Message) RedisCommand {
	var args redis.ArrayMessage

	if message, ok := message.(redis.ArrayMessage); ok{
		args = message
	}else{
		args = []redis.Message{message}
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

func ConvertToMessage(args ...string) redis.Message{
	msg := redis.ArrayMessage{}
	for _, arg := range args{
		msg  = append(msg, redis.NewBulkStringMessage(arg))
	}
	return msg
}

var handlerMap = map[string]Handler{
	"PING": Ping,
	"SET": Set,
	"GET": Get,
}