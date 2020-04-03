package service

import (
	"github.com/liujiangwei/xxcache/protocol"
	"strings"
)

var cmdMap = map[string]Commander{
	"PING": PingCommand{},
	"GET" : GetCommand{},
	"SET": SetCommand{},
}

func convertToHandler(message protocol.Message) (Commander, *protocol.ArrayMessage){
	if message == nil{
		return nil, nil
	}

	var args protocol.ArrayMessage

	if message, ok := message.(protocol.ArrayMessage); ok{
		args = message
	}else{
		args = protocol.ArrayMessage{Data:[]protocol.Message{message}}
	}

	if len(args.Data) == 0{
		return nil,nil
	}

	if cmd, ok := cmdMap[strings.ToUpper(args.Data[0].String())];ok{
		return cmd, &args
	}else{
		return nil, &args
	}
}