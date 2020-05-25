package command

import (
	"github.com/liujiangwei/xxcache/redis"
)

func Set(commander Commander, args redis.ArrayMessage) redis.Message{
	if len(args) != 3{
		return ErrWrongNumberOfArguments("set")
	}

	var key = args[1].String()
	var value = args[2].String()

	commander.Set(key, value)

	return redis.OK
}

func Get(commander Commander, args redis.ArrayMessage) redis.Message{
	if len(args) == 1 || len(args) > 2{
		return ErrWrongNumberOfArguments("get")
	}

	var key = args[1].String()
	if key == ""{
		return ErrWrongNumberOfArguments("get")
	}

	if value, err := commander.Get(key); err != nil{
		if value == ""{
			return redis.ErrorMessage(err.Error())
		}else{
			return redis.Nil
		}
	}else{
		return redis.NewBulkStringMessage(value)
	}
}