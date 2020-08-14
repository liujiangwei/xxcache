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
}


// 定时服务
func (c *Cache) Cron() {
	go c.cronExpire(time.Second * 2)
}

var ErrKeyNil = errors.New("redis nil")
var ErrWrongType = errors.New("wrong type,operation against a key holding the wrong kind of value")
var ErrOffsetOutOfRange = errors.New("offset is out of range")

var OK = "OK"
