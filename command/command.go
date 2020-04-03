package command

import (
	"github.com/liujiangwei/xxcache/protocol"
	"github.com/liujiangwei/xxcache/service"
)

type Commander interface {
	Ping(string) string
}

type PingCommand struct {
	args []protocol.Message
}

func (cmd PingCommand) Exec(commander service.Commander) protocol.Message{
	return protocol.SimpleStringMessage{Data:commander.Ping("a")}
}