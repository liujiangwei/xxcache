package command

import (
	"github.com/liujiangwei/xxcache/redis"
)

//func (server *Server) Get(key database.Key) (database.StringEntry, error){
type Commander interface {
	Ping(string) string
	Set(string, string)
	Get(string) (string, error)
}

type Handler func(commander Commander,args redis.ArrayMessage) redis.Message

type RedisCommand struct {
	args    redis.ArrayMessage
	handler Handler
}

func (command *RedisCommand) Exec(commander Commander) redis.Message{
	return command.handler(commander, command.args)
}


type Command struct {
	Args []string
}

func (cmd Command) Serialize() redis.Message{
	return redis.ConvertToMessage(cmd.Args...)
}

type StringCommand struct {
	Command
}

