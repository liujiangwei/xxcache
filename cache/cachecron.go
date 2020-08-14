package cache

import "time"

// expired key
func (c *Cache) cronExpire(duration time.Duration) {
	for range time.NewTicker(duration).C{
		c.lock.Lock()
		c.dataDict.Range(func(key interface{}, value interface{}) bool {
			expire, ok := value.(time.Time)
			if !ok{
				c.expiresDict.Delete(key)
			}

			if expire.Before(time.Now()){
				c.expiresDict.Delete(key)
				c.dataDict.Delete(key)
			}

			return true
		})

		c.lock.Unlock()
	}
}