package xxcache

import (
	"errors"
	"github.com/liujiangwei/xxcache/list"
)

//CacheValueList for list
type CacheValueList struct {
	List *list.List
}

// ErrListPopFromEmpty you can not pop from an empty
var ErrListPopFromEmpty = errors.New("pop from empty list")

// RPush push values to right
func (cache *XXCache) RPush(key string, value string) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)
	if v == nil {
		v = &CacheValueList{
			List: list.New(),
		}

		cache.set(key, v)
	}

	switch v.(type) {
	case *CacheValueList:
		v := v.(*CacheValueList)

		v.List.InsertPrev(v.List.Tail(), value)

		return nil
	default:
		return ErrKeyType
	}
}

// LPop pop values from left
func (cache *XXCache) LPop(key string) (string, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)
	if v == nil {
		return "", ErrListPopFromEmpty
	}

	switch v.(type) {
	case *CacheValueList:
		v := v.(*CacheValueList)
		n := v.List.Head().Next()
		if !v.List.IsTail(n) {
			v.List.Remove(n)
			return n.Value(), nil
		} else {
			cache.delete(key)
			return "", ErrListPopFromEmpty
		}
	default:
		return "", ErrKeyType

	}
}
