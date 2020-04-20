package cache

import (
	"errors"
	"time"
)

func (cache *Cache) Ping(message string) string{
	if message == ""{
		message = "PONG"
	}

	return message
}

func (cache *Cache) Select(number int) error{
	if number < 0 || number > len(cache.database){
		return errors.New("wrong number")
	}

	cache.databaseSelected = cache.database[number]

	return nil
}

func (cache *Cache) Del(keys ...string) int{
	cache.Lock()
	defer cache.Unlock()

	var num  = 0

	for _, key := range keys{
		if _ , ok := cache.databaseSelected.Get(key); ok{
			num++
			cache.databaseSelected.Delete(key)
		}
	}

	return num
}

func (cache *Cache) Expire(key string, expire time.Duration) int {
	cache.Lock()
	defer cache.Unlock()

	if _, ok  := cache.databaseSelected.Get(key); ok{
		cache.databaseSelected.Expired(key, expire)
		return 1
	}

	return 0
}