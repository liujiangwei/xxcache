package xxcache

import (
	"time"
)

type StringCommand interface {
	Set(key, value string) (string, error)
	SetNX(key, value string) (string, error)
	SetEX(key, value string, expires uint64) (string, error)
	PSetEX(key, value string, expires uint64) (string, error)
	Get(key string) (string, error)
	GetSet(key, value string) (string, error)
	StrLen(key string) (int, error)
	Append(key, value string) (int, error)
	SetRange(key string, pos int, replace string) (int, error)
	GetRange(key string, start, end int) (string, error)
	Incr(key string) (int, error)
	IncrBy(key string) (int, error)
	IncrByFloat(key string, increment float64) (float64, error)
	Decr(key string) (int, error)
	DecrBy(key string) (int, error)
	MSet(kv map[string]string) (string, error)
	MSetNX(kv map[string]string) (string, error)
	MGet(keys ...string) ([]string, error)
}

type ListCommand interface {
	LPush(key string, values ...string) (int, error)
	LPushX(key string, values ...string) (int, error)
	RPush(key string, values ...string) (int, error)
	RPushX(key string, values ...string) (int, error)
	LPop(key string) (string, error)
	RPop(key string) (string, error)
	RPopLPush(keyFrom, keyDestination string) (string, error)
	LRem(key string, count int, value string) (int, error)
	LLen(key string) (int, error)
	LIndex(key string, index int) (string, error)
	LInsert(key string, direction string, pivot string, value string) (int, error)
	LSet(key string, index int, value string) (string, error)
	LRange(key string, start int, stop int) ([]string, error)
	LTrim(key string, start int, stop int) (string, error)
	BLPop(timeout time.Duration, keys ...string)
	BRPop(timeout time.Duration, keys ...string)
	BRPopLPush(keyFrom, keyDestination string, timeout time.Duration)
}

type ExpiresCommand interface {
	Expire(key string, seconds int) (int, error)
	ExpireAt(key string, timestamp time.Time) (int, error)
	Ttl(key string) (int, error)
	// 当生存时间移除成功时，返回 1 . 如果 key 不存在或 key 没有设置生存时间，返回 0 。
	Persist(key string) (int, error)
	PExpire(key string, milliseconds int) (int, error)
	PExpireAt(key string, timestamp time.Time) (int, error)
	PTtl(key string) (int, error)
}

type HashCommand interface {
	HSet()
	HSetNX()
	HGet()
	HExists()
	HDel()
	HLen()
	HStrLen()
	HIncrBy()
	HIncrByFloat()
	HMSet()
	HMGet()
	HKeys()
	HValues()
	HGetAll()
	HScan()
}

func (client Client) Get(key string) (string, error) {
	return client.cache.Get(key)
}

func (client Client) Set(key string, value string) (string, error){
	return client.redis.Set(key, value)
}

