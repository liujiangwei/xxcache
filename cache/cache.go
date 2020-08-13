package cache

import (
	"errors"
	"github.com/cornelk/hashmap"
	"sync"
	"time"
)

type Cache struct {
	dataDict    hashmap.HashMap
	expiresDict hashmap.HashMap
	lock        sync.Mutex
}

func (c *Cache)Del(key string) int{
	if _, err := c.Get(key); err != nil{
		return 0
	}else{
		c.dataDict.Del(key)
		c.expiresDict.Del(key)
	}

	return 1
}

func (c *Cache) LPush(key string, values ...string) (int, error) {
	panic("implement me")
}

func (c *Cache) LPushX(key string, values ...string) (int, error) {
	panic("implement me")
}

func (c *Cache) RPush(key string, values ...string) (int, error) {
	panic("implement me")
}

func (c *Cache) RPushX(key string, values ...string) (int, error) {
	panic("implement me")
}

func (c *Cache) LPop(key string) (string, error) {
	panic("implement me")
}

func (c *Cache) RPop(key string) (string, error) {
	panic("implement me")
}

func (c *Cache) RPopLPush(keyFrom, keyDestination string) (string, error) {
	panic("implement me")
}

func (c *Cache) LRem(key string, count int, value string) (int, error) {
	panic("implement me")
}

func (c *Cache) LLen(key string) (int, error) {
	panic("implement me")
}

func (c *Cache) LIndex(key string, index int) (string, error) {
	panic("implement me")
}

func (c *Cache) LInsert(key string, direction string, pivot string, value string) (int, error) {
	panic("implement me")
}

func (c *Cache) LSet(key string, index int, value string) (string, error) {
	panic("implement me")
}

func (c *Cache) LRange(key string, start int, stop int) ([]string, error) {
	panic("implement me")
}

func (c *Cache) LTrim(key string, start int, stop int) (string, error) {
	panic("implement me")
}

func (c *Cache) BLPop(timeout time.Duration, keys ...string) {
	panic("implement me")
}

func (c *Cache) BRPop(timeout time.Duration, keys ...string) {
	panic("implement me")
}

func (c *Cache) BRPopLPush(keyFrom, keyDestination string, timeout time.Duration) {
	panic("implement me")
}

var ErrKeyNil = errors.New("redis nil")
var ErrWrongType = errors.New("wrong type,operation against a key holding the wrong kind of value")
var ErrOffsetOutOfRange = errors.New("offset is out of range")
var ErrIntegerOrOutOfRange = errors.New("value is not an integer or out of range")

var OK = "OK"
