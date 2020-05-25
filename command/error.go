package command

import (
	"github.com/liujiangwei/xxcache/redis"
)

func ErrWrongNumberOfArguments(cmd string) redis.ErrorMessage {
	return redis.ErrorMessage("ERR wrong number of arguments for '" + cmd + "' command")
}
