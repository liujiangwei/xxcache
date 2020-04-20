package command

import (
	"github.com/liujiangwei/xxcache/rconn"
)

func ErrWrongNumberOfArguments(cmd string) rconn.ErrorMessage {
	return rconn.ErrorMessage("ERR wrong number of arguments for '" + cmd + "' command")
}
