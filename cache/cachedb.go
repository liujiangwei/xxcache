package cache

func (c *Cache)Del(key string) (n int){
	c.lock.Lock()
	defer c.lock.Unlock()

	if _,ok := c.get(key); ok{
		n =1
	}

	c.dataDict.Delete(key)
	c.expiresDict.Delete(key)

	return n
}

func (c *Cache)Flush() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.dataDict = &SyncMapDatabase{}
	c.expiresDict = &SyncMapDatabase{}
}