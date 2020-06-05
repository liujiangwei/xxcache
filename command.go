package xxcache

import "github.com/liujiangwei/xxcache/database"

type RedisStringCommand interface {
	Set(key,value string) (string, error)
	SetNX(key, value string)(string, error)
	SetEX(key, value string, expires uint64)(string, error)
	PSetEX(key, value string, expires uint64)(string, error)
	GET(key string)(string, error)
	GetSet(key, value string)(string, error)
	StrLen(key string)(int, error)
	Append(key string)(int, error)
	SetRange(key string, pos int, replace string)(int, error)
	GetRange(key string, start, end int)(string, error)
	Incr(key string)(int, error)
	IncrBy(key string)(int, error)
	IncrByFloat(key string, increment float64)(float64, error)
	Decr(key string)(int, error)
	DecrBy(key string)(int, error)
	MSet(kv map[string]string)(string, error)
	MSetNX(kv map[string]string)(string, error)
	MGet(keys ...string)([]string, error)
}


type CacheCommand interface {
	Key() string
	Entry(entry database.Entry)
}

type KeyCommand struct {
	key string
}

func (cmd KeyCommand) Key() string {
	return cmd.key
}

func NewKeyCommand(key string) KeyCommand {
	return KeyCommand{key:key}
}

//
type CacheStringCommand struct {
	KeyCommand
	entry *database.StringEntry
}

// search key entry in local cache
func (cmd CacheStringCommand) Entry(entry database.Entry) {
	if entry, ok := entry.(*database.StringEntry); ok {
		cmd.entry = entry
	}
}