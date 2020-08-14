package cache

type HashEntry struct {
	SyncMapDatabase
}

func (he *HashEntry) Get(field string) (value string, ok bool) {
	v, ok := he.Load(field)
	if !ok {
		return "", ok
	}

	// here is must string
	value, _ = fmtString(v)
	return value, ok
}

func (he *HashEntry) Set(field string, value string) (n int) {
	v, ok := he.Get(field)
	if !ok {
		he.Store(field, tryInt64(value))
		return 1
	}

	if v != value {
		he.Store(field, tryInt64(value))
		n = 1
	}

	return n
}

func newHash() *HashEntry {
	return &HashEntry{}
}

func (c *Cache) getHashEntry(key string) (*HashEntry, error) {
	entry, ok := c.get(key)
	if !ok {
		return nil, nil
	}

	if he, ok := entry.(*HashEntry); ok {
		return he, nil
	} else {
		return nil, ErrWrongType
	}
}

func (c *Cache) HSet(key string, field string, value string) (n int, err error) {
	he, err := c.getHashEntry(key)
	if err != nil {
		return 0, err
	}

	if he == nil {
		he = newHash()
		c.dataDict.Store(key, he)
	}

	return he.Set(field, value), nil
}

func (c *Cache) HGet(key string, field string) (value string, err error) {
	he, err := c.getHashEntry(key)
	if err != nil {
		return value, err
	}

	// key is not set
	if he == nil {
		return value, ErrKeyNil
	}

	// field is not set
	if str, ok := he.Get(field); !ok {
		err = ErrKeyNil
	} else {
		value = str
	}

	return value, err
}
