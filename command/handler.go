package command

import (
	"github.com/liujiangwei/xxcache/redis"
)

func notFound(commander Commander, args redis.ArrayMessage) redis.Message {
	if len(args) >= 1 {
		return ErrWrongNumberOfArguments(args[0].String())
	}

	return ErrWrongNumberOfArguments("")
}

func Ping(commander Commander, args redis.ArrayMessage) redis.Message {
	var message string

	if len(args) > 2 {
		return ErrWrongNumberOfArguments("Ping")
	}

	if len(args) == 2 {
		message = args[1].String()
	}

	var reply = commander.Ping(message)
	if message == "" {
		return redis.SimpleStringMessage(reply)
	} else {
		return redis.NewBulkStringMessage(reply)
	}

}
