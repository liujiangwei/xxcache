package cache

import (
	"errors"
	"time"
)

type Cache struct {
	Database Database
}

func (cache *Cache) get(key string) interface{} {
	return cache.Database.Get(key)
}

func (cache *Cache) expire(key string, seconds int)  {

}

func (cache *Cache) LPush(key string, values ...string) (int, error) {
	panic("implement me")
}

func (cache *Cache) LPushX(key string, values ...string) (int, error) {
	var entry *ListEntry
	e := cache.Database.Get(key)
	switch e.(type) {
	case nil:
		entry =  new(ListEntry)
		cache.Database.SetList(key, entry)
	case *ListEntry:
		entry = e.(*ListEntry)
	default:
		return 0, ErrWrongType
	}

	for _, value := range values{
		entry.AppendHead(value)
	}
	return len(values), nil
}

func (cache *Cache) RPush(key string, values ...string) (int, error) {
	var entry *ListEntry
	e := cache.Database.Get(key)
	switch e.(type) {
	case nil:
		entry =  new(ListEntry)
		cache.Database.SetList(key, entry)
	case *ListEntry:
		entry = e.(*ListEntry)
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
var ErrOffsetOutOfRange = errors.New("offset is out of range")
var ErrIntegerOrOutOfRange =errors.New("value is not an integer or out of range")

var OK = "OK"
