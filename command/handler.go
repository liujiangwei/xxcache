package command

import "github.com/liujiangwei/xxcache/protocol"

func notFound(commander Commander, args *protocol.ArrayMessage) protocol.Message{
	if len(args.Data) >= 1{
		return ErrWrongNumberOfArguments(args.Data[0].String())
	}

	return ErrWrongNumberOfArguments("")
}

func Ping(commander Commander, args *protocol.ArrayMessage) protocol.Message{
	var message string

	if len(args.Data) > 2{
		return ErrWrongNumberOfArguments("Ping")
	}

	if len(args.Data) == 2{
		message = args.Data[1].String()
	}

	return protocol.SimpleStringMessage{Data:commander.Ping(message)}
}

