package command

import (
	"github.com/liujiangwei/xxcache/protocol"
)

type Commander interface {
	Ping(string) string
	Set(string, string) (string, error)
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