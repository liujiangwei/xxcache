package command

import (
	"github.com/liujiangwei/xxcache/protocol"
)

//func (server *Server) Get(key entry.Key) (entry.StringEntry, error){
type Commander interface {
	Ping(string) string
	Set(string, string)
	Get(string) (string, error)
}

type Handler func(commander Commander,args *protocol.ArrayMessage) protocol.Message

type RedisCommand struct {
	args    *protocol.ArrayMessage
	handler Handler
}

func (command *RedisCommand) Exec(commander Commander) protocol.Message{
	return command.handler(commander, command.args)
}