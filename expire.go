package xxcache

import "time"

func (cache *XXCache) checkExpire() {
	ticker := time.NewTicker(time.Second)
	for {
		<-ticker.C
		cache.lock.Lock()
		for k, t := range cache.expires {
			if t.Before(time.Now()) {
				cache.delete(k)
			}
		}
		cache.lock.Unlock()
	}
}
