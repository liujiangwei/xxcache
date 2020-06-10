package cache

import (
	"errors"
	"github.com/liujiangwei/xxcache/cache/database"
	"time"
)

type Cache struct {
	Database database.Database
}
func (cache *Cache) LPush(key string, values ...string) (int, error) {
	panic("implement me")
}

func (cache *Cache) LPushX(key string, values ...string) (int, error) {
	var entry *database.ListEntry
	e := cache.Database.Get(key)
	switch e.(type) {
	case nil:
		entry =  new(database.ListEntry)
		cache.Database.SetList(key, entry)
	case *database.ListEntry:
		entry = e.(*database.ListEntry)
	default:
		return 0, ErrWrongType
	}

	for _, value := range values{
		entry.AppendHead(value)
	}
	return len(values), nil
}

func (cache *Cache) RPush(key string, values ...string) (int, error) {
	var entry *database.ListEntry
	e := cache.Database.Get(key)
	switch e.(type) {
	case nil:
		entry =  new(database.ListEntry)
		cache.Database.SetList(key, entry)
	case *database.ListEntry:
		entry = e.(*database.ListEntry)
	default:
		return 0, ErrWrongType
	}

	for _, value := range values{
		entry.AppendTail(value)
	}

	return len(values), nil
}

func (cache *Cache) RPushX(key string, values ...string) (int, error) {
	panic("implement me")
}

func (cache *Cache) LPop(key string) (string, error) {
	panic("implement me")
}

func (cache *Cache) RPop(key string) (string, error) {
	panic("implement me")
}

func (cache *Cache) RPopLPush(keyFrom, keyDestination string) (string, error) {
	panic("implement me")
}

func (cache *Cache) LRem(key string, count int, value string) (int, error) {
	panic("implement me")
}

func (cache *Cache) LLen(key string) (int, error) {
	panic("implement me")
}

func (cache *Cache) LIndex(key string, index int) (string, error) {
	panic("implement me")
}

func (cache *Cache) LInsert(key string, direction string, pivot string, value string) (int, error) {
	panic("implement me")
}

func (cache *Cache) LSet(key string, index int, value string) (string, error) {
	panic("implement me")
}

func (cache *Cache) LRange(key string, start int, stop int) ([]string, error) {
	panic("implement me")
}

func (cache *Cache) LTrim(key string, start int, stop int) (string, error) {
	panic("implement me")
}

func (cache *Cache) BLPop(timeout time.Duration, keys ...string) {
	panic("implement me")
}

func (cache *Cache) BRPop(timeout time.Duration, keys ...string) {
	panic("implement me")
}

func (cache *Cache) BRPopLPush(keyFrom, keyDestination string, timeout time.Duration) {
	panic("implement me")
}

var ErrKeyNil = errors.New("redis nil")
var ErrWrongType =  errors.New("wrong type,operation against a key holding the wrong kind of value")
var OK = "OK"

func (cache *Cache) Set(key, value string) (string, error) {
	var entry *database.StringEntry

	if e, ok := cache.Database.Get(key).(*database.StringEntry); !ok{
		entry = &database.StringEntry{}
		cache.Database.SetString(key, entry)
	}else{
		entry = e
	}

	entry.Set(value)

	return OK, nil
}

func (cache *Cache) SetNX(key, value string) (string, error) {
	panic("implement me")
}

func (cache *Cache) SetEX(key, value string, expires uint64) (string, error) {
	panic("implement me")
}

func (cache *Cache) PSetEX(key, value string, expires uint64) (string, error) {
	panic("implement me")
}

func (cache *Cache) Get(key string) (string, error) {
	var entry = cache.Database.Get(key)
	if entry == nil{
		return "", ErrKeyNil
	}

	if entry, ok := entry.(*database.StringEntry); ok{
		return entry.Get(), nil
	}else{
		return "", ErrWrongType
	}
}

func (cache *Cache) GetSet(key, value string) (string, error) {
	var entry *database.StringEntry
	var oldVal string
	e := cache.Database.Get(key)
	switch e.(type) {
	case nil:
		entry = &database.StringEntry{}
		cache.Database.SetString(key, entry)
	case *database.StringEntry:
		entry = e.(*database.StringEntry)
	default:
		return oldVal, ErrWrongType
	}
	oldVal = entry.Get()
	entry.Set(value)

	return oldVal, nil
}

func (cache *Cache) StrLen(key string) (int, error) {
	var entry = cache.Database.Get(key)
	if entry == nil{
		return 0, nil
	}

	if entry, ok := entry.(*database.StringEntry); ok{
		return len(entry.Get()), nil
	}else{
		return 0, ErrWrongType
	}
}

func (cache *Cache) Append(key string, value string) (int, error) {
	var entry *database.StringEntry
	e := cache.Database.Get(key)
	switch e.(type) {
	case nil:
		entry = &database.StringEntry{}
		cache.Database.SetString(key, entry)
	case *database.StringEntry:
		entry = e.(*database.StringEntry)
	default:
		return 0, ErrWrongType
	}

	entry.Set(entry.Get() + value)

	return len(entry.Get()), nil
}

func (cache *Cache) SetRange(key string, pos int, replace string) (int, error) {
	panic("implement me")
}

func (cache *Cache) GetRange(key string, start, end int) (string, error) {
	panic("implement me")
}

func (cache *Cache) Incr(key string) (int, error) {
	panic("implement me")
}

func (cache *Cache) IncrBy(key string) (int, error) {
	panic("implement me")
}

func (cache *Cache) IncrByFloat(key string, increment float64) (float64, error) {
	panic("implement me")
}

func (cache *Cache) Decr(key string) (int, error) {
	panic("implement me")
}

func (cache *Cache) DecrBy(key string) (int, error) {
	panic("implement me")
}

func (cache *Cache) MSet(kv map[string]string) (string, error) {
	panic("implement me")
}

func (cache *Cache) MSetNX(kv map[string]string) (string, error) {
	panic("implement me")
}

func (cache *Cache) MGet(keys ...string) ([]string, error) {
	panic("implement me")
}
