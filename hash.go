package xxcache

import "errors"

type CacheValueHash struct {
	hash   map[string]string
	length int
}

var ErrHashFiledNotFound = errors.New("hash field not found")
var ErrHashFieldExist = errors.New("hash field has already exist")

func (cache *XXCache) HSet(key, field, value string) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)
	if v == nil {
		v = &CacheValueHash{
			hash: make(map[string]string),
		}

		cache.set(key, v)
	}

	if v, ok := v.(*CacheValueHash); ok {
		v.hash[field] = value
		v.length++
		return nil
	} else {
		return ErrKeyType
	}
}

func (cache *XXCache) HGet(key, field string) (string, error) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)

	if v == nil {
		return "", ErrKeyNotExist
	} else if v, ok := v.(*CacheValueHash); ok {
		if value, found := v.hash[field]; found {
			return value, nil
		} else {
			return "", ErrHashFiledNotFound
		}
	} else {
		return "", ErrKeyType
	}
}

func (cache *XXCache) HDel(key string, fields ...string) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)

	if v == nil {
		return ErrKeyNotExist
	}

	if v, ok := v.(*CacheValueHash); ok {
		for _, field := range fields {
			delete(v.hash, field)
			v.length--
		}

		if v.length == 0 {
			cache.delete(key)
		}

		return nil
	} else {
		return ErrKeyType
	}
}

func (cache *XXCache) HSetNX(key, field, value string) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	v := cache.get(key)
	if v == nil {
		v = &CacheValueHash{
			hash: make(map[string]string),
		}

		cache.set(key, v)
	}

	if v, ok := v.(*CacheValueHash); ok {
		if _, found := v.hash[field]; found {
			return ErrHashFieldExist
		} else {
			v.hash[field] = value
			v.length++
			return nil
		}
	} else {
		return ErrKeyType
	}
}
