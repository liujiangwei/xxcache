package command

import "github.com/liujiangwei/xxcache/protocol"

func ErrWrongNumberOfArguments(cmd string) protocol.ErrorMessage {
	return protocol.ErrorMessage{
		Data: "ERR wrong number of arguments for '" + cmd + "' command",
	}
}