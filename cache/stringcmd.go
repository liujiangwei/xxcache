package cache

import (
	"errors"
	"strconv"
)

//WRONGTYPE Operation against a key holding the wrong kind of value
var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

// "", error
// "nil", error
// "", nil
// "a",nil
func (cache *Cache) Get(key string) (string, error) {
	cache.Lock()
	defer cache.Unlock()

	v, ok := cache.databaseSelected.Get(key)
	if !ok {
		return Nil, ErrKeyNotExist
	}
	if v, ok := v.(string); ok {
		return v, nil
	} else {
		return "", ErrWrongType
	}
}

func (cache *Cache) Set(key string, value string) {
	cache.Lock()
	defer cache.Unlock()

	cache.databaseSelected.Set(key, value)
	return
}

func (cache *Cache) Incr(key string) (int, error) {
	cache.Lock()
	defer cache.Unlock()

	v, ok := cache.databaseSelected.Get(key)
	if !ok {
		cache.databaseSelected.Set(key, "1")
		return 1, nil
	}

	s, ok := v.(string)
	if !ok {
		return 0, errors.New("ERR value is not an integer or out of range")
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.New("ERR value is not an integer or out of range," + err.Error())
	}

	cache.databaseSelected.Set(key, strconv.Itoa(i+1))

	return i + 1, nil
}

var Nil = "Nil"

var ErrKeyNotExist = errors.New("error key is not exists")
