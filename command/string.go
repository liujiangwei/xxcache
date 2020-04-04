package command

import "github.com/liujiangwei/xxcache/protocol"

func Set(commander Commander, args *protocol.ArrayMessage) protocol.Message{
	if len(args.Data) != 3{
		return ErrWrongNumberOfArguments("set")
	}

	var key = args.Data[1].String()
	var value = args.Data[2].String()

	if ok, err := commander.Set(key, value); err == nil{
		return protocol.SimpleStringMessage{Data:ok}
	}else{
		return protocol.ErrorMessage{Data:err.Error()}
	}
}

func Get(commander Commander, args *protocol.ArrayMessage) protocol.Message{
	if len(args.Data) == 1 || len(args.Data) > 2{
		return ErrWrongNumberOfArguments("get")
	}

	var key = args.Data[1].String()
	if key == ""{
		return ErrWrongNumberOfArguments("get")
	}

	if value, err := commander.Get(key); err != nil{
		return protocol.ErrorMessage{Data:err.Error()}
	}else{
		return protocol.SimpleStringMessage{Data:value}
	}
}