package cache

func (cache *Cache)SetAny(key string, value interface{}) {
	cache.Lock()
	defer cache.Unlock()

	cache.databaseSelected.Set(key, value)
}

func (cache *Cache)GetAny(key string) (interface{}, bool){
	cache.Lock()
	defer cache.Unlock()

	return cache.databaseSelected.Get(key)
}

