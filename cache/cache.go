package cache

import (
	"errors"
	"sync"
	"time"
)

func New() *Cache{
	cache :=  Cache{
		dataDict:&SyncMapDatabase{},
		expiresDict:&SyncMapDatabase{},
	}

	return &cache
}

type Cache struct {
	dataDict    Database
	expiresDict Database
	lock        sync.RWMutex

	expiredOnce int
	size int
}

func (c *Cache)Del(key string) (n int){
	c.lock.Lock()
	defer c.lock.Unlock()

	if _,ok := c.get(key); ok{
		n =1
	}

	c.dataDict.Delete(key)
	c.expiresDict.Delete(key)

	return n
}

// 定时服务
func (c *Cache) Cron() {
	go c.cronExpire(time.Second * 2)
}


// expired key
func (c *Cache) cronExpire(duration time.Duration) {
	for range time.NewTicker(duration).C{
		c.lock.Lock()
		c.dataDict.Range(func(key interface{}, value interface{}) bool {
			expire, ok := value.(time.Time)
			if !ok{
				c.expiresDict.Delete(key)
			}

			if expire.Before(time.Now()){
				c.expiresDict.Delete(key)
				c.dataDict.Delete(key)
			}

			return true
		})

		c.lock.Unlock()
	}
}


func (c *Cache)Flush() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.dataDict = &SyncMapDatabase{}
	c.expiresDict = &SyncMapDatabase{}
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
