package xxcache

import (
	"errors"
	"sync"
	"time"
)

//ErrKeyNotExist return when key not found in database
var ErrKeyNotExist = errors.New("key not found")
var ErrKeyType = errors.New("the key type is wrong")
var ErrKeyExist = errors.New("key is exist")

//CacheValue database key => value
type CacheValue struct {
}

// XXCache struct
type XXCache struct {
	lock     sync.Mutex
	database map[string]interface{}
	expires  map[string]time.Time
}

func (cache *XXCache) get(key string) interface{} {
	if v, found := cache.database[key]; !found {
		return nil
	} else if expire, found := cache.expires[key]; found && expire.Before(time.Now()) {
		return nil
	} else {
		return v
	}
}

func (cache *XXCache) set(key string, value interface{}) {
	cache.database[key] = value
}

func (cache *XXCache) expire(key string, expire time.Duration) {
	cache.expires[key] = time.Now().Add(expire)
}

func (cache *XXCache) setWithExpire(key string, value interface{}, expire time.Duration) {
	cache.set(key, value)
	cache.expire(key, expire)
}

func (cache *XXCache) delete(key string) {
	delete(cache.database, key)
	delete(cache.expires, key)
}

func NewXXCache() *XXCache {
	cache := &XXCache{
		database: make(map[string]interface{}),
		expires:  make(map[string]time.Time),
	}

	//go cache.checkExpire()

	return cache
}

// Expire set a checkExpire time to the key
func (cache *XXCache) Expire(key string, expire time.Duration) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)
	if v == nil {
		return ErrKeyNotExist
	}

	cache.expire(key, expire)

	return nil
}

// Del delete
func (cache *XXCache) Del(key string) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)
	if v == nil {
		return ErrKeyNotExist
	}
	cache.delete(key)

	return nil
}
