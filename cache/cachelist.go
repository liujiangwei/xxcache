package cache

import "time"

type ListEntry struct {
	val []string
}

func (entry *ListEntry) AppendTail(val string) {
	entry.val = append(entry.val, val)
}

func (entry *ListEntry) AppendHead(val string) {
	entry.val = append([]string{val}, entry.val...)
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