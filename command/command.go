package command

import (
	"github.com/liujiangwei/xxcache/rconn"
)

//func (server *Server) Get(key entry.Key) (entry.StringEntry, error){
type Commander interface {
	Ping(string) string
	Set(string, string)
	Get(string) (string, error)
}

type Handler func(commander Commander,args rconn.ArrayMessage) rconn.Message

type RedisCommand struct {
	args    rconn.ArrayMessage
	handler Handler
}

func (command *RedisCommand) Exec(commander Commander) rconn.Message{
	return command.handler(commander, command.args)
}