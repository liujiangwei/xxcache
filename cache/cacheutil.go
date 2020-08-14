package cache

import "time"

// if key is expired
func (c *Cache) expired(key string) bool {
	v, ok := c.expiresDict.Load(key)
	if !ok || v == nil {
		return false
	}

	var expired bool
	if expires, ok := v.(time.Time); ok {
		expired = expires.Before(time.Now())
	}

	// delete expired key value
	if expired {
		c.dataDict.Delete(key)
		c.expiresDict.Delete(key)
	}

	return expired
}

// set expire time after duration from now
// if duration < 0, remove expire time
func (c *Cache) expires(key string, duration time.Duration) {
	if duration <= 0{
		c.expiresDict.Delete(key)
	}else{
		c.expiresDict.Store(key, time.Now().Add(duration))
	}
}

func (c *Cache) get(key string) (value interface{},ok bool){
	val, ok := c.dataDict.Load(key)
	if !ok{
		return val, ok
	}

	if c.expired(key) {
		val = nil
		ok = false
	}

	return val, ok
}