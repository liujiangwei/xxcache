package command

import (
	"github.com/liujiangwei/xxcache/rconn"
)

func notFound(commander Commander, args rconn.ArrayMessage) rconn.Message{
	if len(args) >= 1{
		return ErrWrongNumberOfArguments(args[0].String())
	}

	return ErrWrongNumberOfArguments("")
}

func Ping(commander Commander, args rconn.ArrayMessage) rconn.Message{
	var message string

	if len(args) > 2{
		return ErrWrongNumberOfArguments("Ping")
	}

	if len(args) == 2{
		message = args[1].String()
	}

	var reply = commander.Ping(message)
	if message == ""{
		return rconn.SimpleStringMessage(reply)
	}else{
		return rconn.NewBulkStringMessage(reply)
	}

}

