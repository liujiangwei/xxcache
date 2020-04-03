package service

import (
	"github.com/liujiangwei/xxcache/protocol"
)

//ERR wrong number of arguments for 'get' command
func newErrCommandWrongNumberOfArguments(command string) protocol.ErrorMessage{
	return protocol.ErrorMessage{Data:"ERR wrong number of arguments for '"+command+"' command"}
}

type Commander interface {
	Exec(server *Server,args *protocol.ArrayMessage) protocol.Message
}

type PingCommand struct {
}
func (cmd PingCommand) Exec(server *Server, args *protocol.ArrayMessage) protocol.Message{
	var message string

	if len(args.Data) > 2{
		return newErrCommandWrongNumberOfArguments("ping")
	}

	if len(args.Data) == 2{
		message = args.Data[1].String()
	}

	return protocol.SimpleStringMessage{Data:server.Ping(message)}
}

type GetCommand struct {
}

func (cmd GetCommand) Exec(server *Server, args *protocol.ArrayMessage) protocol.Message{
	if len(args.Data) == 1 || len(args.Data) > 2{
		return newErrCommandWrongNumberOfArguments("get")
	}

	var key = args.Data[1].String()
	if key == ""{
		return newErrCommandWrongNumberOfArguments("get")
	}

	return protocol.SimpleStringMessage{Data:server.Get(key)}
}

type SetCommand struct {
}

func (cmd SetCommand) Exec(server *Server, args *protocol.ArrayMessage) protocol.Message{
	if len(args.Data) != 3{
		return newErrCommandWrongNumberOfArguments("set")
	}

	var key = args.Data[1].String()
	var value = args.Data[2].String()

	return protocol.SimpleStringMessage{Data:server.Set(key, value)}
}