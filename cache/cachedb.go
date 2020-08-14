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

func (c *Cache) Flush() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.dataDict = &SyncMapDatabase{}
	c.expiresDict = &SyncMapDatabase{}
}

func (c *Cache) Exists(keys ...string) (n int){
	for _, key := range keys{
		if _, ok := c.dataDict.Load(key); ok{
			n++
		}
	}

	return n
}

func (c *Cache)Type(key string) string {
	v, ok := c.get(key)
	if !ok{
		return "none"
	}

	switch v.(type) {
	case hashEntry:
		return "hash"
	default:
		return "string"
	}
}