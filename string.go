package xxcache

import (
	"log"
	"time"
)

// CacheValueString values of string type
type CacheValueString struct {
	value string
}

// Get values from a string type key
func (cache *XXCache) Get(key string) (string, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)
	if v == nil{
		return "", ErrKeyNotExist
	}

	log.Println(v)
	if v,ok := v.(*CacheValueString); ok {
		return v.value, nil
	}else {
		return "", ErrKeyType
	}
}

//Set set the value of a key
func (cache *XXCache) Set(key, value string) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)
	if v == nil{
		v = &CacheValueString{value:value}
		cache.set(key, v)
	}

	if v,ok := v.(*CacheValueString); ok{
		v.value = value
		return nil
	}else{
		return ErrKeyType
	}
}

//SetEX set the value and expiration of a key
func (cache *XXCache) SetEX(key, value string, expire time.Duration) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)
	if v == nil{
		v = &CacheValueString{value: value}
		cache.set(key, v)
	}

	if v, ok := v.(*CacheValueString); ok{
		v.value = value
		cache.expire(key, expire)
		return nil
	}else{
		return ErrKeyType
	}
}

// SetEN set the value of a key, only if the key was not exist
func (cache *XXCache) SetEN(key, value string) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)

	if v == nil{
		v = &CacheValueString{value: value}
	}

	if v, ok := v.(*CacheValueString); ok{
		v.value = value
		cache.set(key, &v)
		return nil
	}else{
		return ErrKeyType
	}
}
