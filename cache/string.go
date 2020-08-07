package cache

type StringEntry struct {
	data string
}

func (entry *StringEntry)Get() string{
	return entry.data
}

func (entry *StringEntry) Set(val string) {
	entry.data = val
}



func (cache *Cache) Set(key, value string) (string, error) {
	var entry *StringEntry

	if e, ok := cache.Database.Get(key).(*StringEntry); !ok{
		entry = &StringEntry{}
		cache.Database.SetString(key, entry)
	}else{
		entry = e
	}

	entry.Set(value)

	return OK, nil
}

func (cache *Cache) SetNX(key, value string) (string, error) {
	panic("implement me")
}

func (cache *Cache) SetEX(key, value string, expires uint64) (string, error) {
	panic("implement me")
}

func (cache *Cache) PSetEX(key, value string, expires uint64) (string, error) {
	panic("implement me")
}

func (cache *Cache) Get(key string) (string, error) {
	var entry = cache.Database.Get(key)
	if entry == nil{
		return "", ErrKeyNil
	}

	if entry, ok := entry.(*StringEntry); ok{
		return entry.Get(), nil
	}else{
		return "", ErrWrongType
	}
}

func (cache *Cache) GetSet(key, value string) (string, error) {
	var entry *StringEntry
	var oldVal string
	e := cache.Database.Get(key)
	switch e.(type) {
	case nil:
		entry = &StringEntry{}
		cache.Database.SetString(key, entry)
	case *StringEntry:
		entry = e.(*StringEntry)
	default:
		return oldVal, ErrWrongType
	}
	oldVal = entry.Get()
	entry.Set(value)

	return oldVal, nil
}

func (cache *Cache) StrLen(key string) (int, error) {
	var entry = cache.Database.Get(key)
	if entry == nil{
		return 0, nil
	}

	if entry, ok := entry.(*StringEntry); ok{
		return len(entry.Get()), nil
	}else{
		return 0, ErrWrongType
	}
}

func (cache *Cache) Append(key string, value string) (int, error) {
	var entry *StringEntry
	e := cache.Database.Get(key)
	switch e.(type) {
	case nil:
		entry = &StringEntry{}
		cache.Database.SetString(key, entry)
	case *StringEntry:
		entry = e.(*StringEntry)
	default:
		return 0, ErrWrongType
	}

	entry.Set(entry.Get() + value)

	return len(entry.Get()), nil
}

func (cache *Cache) SetRange(key string, pos int, replace string) (int, error) {
	panic("implement me")
}

func (cache *Cache) GetRange(key string, start, end int) (string, error) {
	panic("implement me")
}

func (cache *Cache) Incr(key string) (int, error) {
	panic("implement me")
}

func (cache *Cache) IncrBy(key string) (int, error) {
	panic("implement me")
}

func (cache *Cache) IncrByFloat(key string, increment float64) (float64, error) {
	panic("implement me")
}

func (cache *Cache) Decr(key string) (int, error) {
	panic("implement me")
}

func (cache *Cache) DecrBy(key string) (int, error) {
	panic("implement me")
}

func (cache *Cache) MSet(kv map[string]string) (string, error) {
	panic("implement me")
}

func (cache *Cache) MSetNX(kv map[string]string) (string, error) {
	panic("implement me")
}

func (cache *Cache) MGet(keys ...string) ([]string, error) {
	panic("implement me")
}
