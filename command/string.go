package command

import (
	"github.com/liujiangwei/xxcache/rconn"
)

func Set(commander Commander, args rconn.ArrayMessage) rconn.Message{
	if len(args) != 3{
		return ErrWrongNumberOfArguments("set")
	}

	var key = args[1].String()
	var value = args[2].String()

	commander.Set(key, value)

	return rconn.OK
}

func Get(commander Commander, args rconn.ArrayMessage) rconn.Message{
	if len(args) == 1 || len(args) > 2{
		return ErrWrongNumberOfArguments("get")
	}

	var key = args[1].String()
	if key == ""{
		return ErrWrongNumberOfArguments("get")
	}

	if value, err := commander.Get(key); err != nil{
		if value == ""{
			return rconn.ErrorMessage(err.Error())
		}else{
			return rconn.Nil
		}
	}else{
		return rconn.NewBulkStringMessage(value)
	}
}