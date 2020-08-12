package cache

import (
	"strconv"
	"sync/atomic"
	"time"
)

type StringEntry struct {
	data interface{}
}

func (entry *StringEntry) Value() string{
	switch entry.data.(type) {
	case *int64:
		return strconv.Itoa(int(*entry.data.(*int64)))
	case *string:
		return entry.data.(string)
	default:
		return ""
	}
}

func (entry *StringEntry) SetValue(val string) {
	if n,err := strconv.Atoi(val); err != nil{
		n := int64(n)
		entry.data = &n
	}else{
		s := val
		entry.data = &s
	}
}

func (entry *StringEntry) Incr(delta int64) (int64, bool){
	switch entry.data.(type) {
	case *int64:
		n := atomic.AddInt64(entry.data.(*int64),delta)
		return n, true
	default:
		return 0, false
	}
}

func (cache *Cache) Set(key, value string) (string, error) {
	cache.Database.dataDict.Cas()
	entry := cache.get(key)
	switch entry.(type) {
	case *StringEntry:
		entry.(*StringEntry).SetValue(value)
	default:
		entry := new(StringEntry)
		entry.SetValue(value)
		cache.Database.SetString(key, entry)
	}

	return OK, nil
}

// SET if Not eXists
func (cache *Cache) SetNX(key, value string) int {
	entry := cache.get(key)

	if entry == nil{
		entry := new(StringEntry)
		entry.SetValue(value)
		cache.Database.SetString(key, entry)
		return 1
	}

	return 0
}

// SETEX key seconds value
func (cache *Cache) SetEX(key, value string, expires int64) (string, error) {
	entry := cache.get(key)
	switch entry.(type) {
	case *StringEntry:
		entry.(*StringEntry).SetValue(value)
	default:
		entry := new(StringEntry)
		entry.SetValue(value)
		cache.Database.SetString(key, entry)
	}

	cache.Database.Expires(key, time.Second * time.Duration(expires))

	return OK, nil
}

func (cache *Cache) PSetEX(key, value string, expires int64) (string, error) {
	entry := cache.get(key)
	switch entry.(type) {
	case *StringEntry:
		entry.(*StringEntry).SetValue(value)
	default:
		entry := new(StringEntry)
		entry.SetValue(value)
		cache.Database.SetString(key, entry)
	}

	cache.Database.Expires(key, time.Millisecond * time.Duration(expires))

	return OK, nil
}

func (cache *Cache) Get(key string) (string, error) {
	var entry = cache.get(key)
	switch  entry.(type){
	case nil:
		return "", ErrKeyNil
	case *StringEntry:
		return entry.(*StringEntry).Value(), nil
	default:
		return "", ErrWrongType
	}
}

func (cache *Cache) GetSet(key, value string) (string, error) {
	var oldVal string
	entry := cache.get(key)
	switch entry.(type) {
	case nil:
		return "", ErrKeyNil
	case *StringEntry:
		oldVal = entry.(*StringEntry).Value()
		entry.(*StringEntry).SetValue(value)
		return oldVal, nil
	default:
		return "", ErrWrongType
	}
}

func (cache *Cache) StrLen(key string) (int, error) {
	var entry = cache.get(key)
	switch entry.(type) {
	case nil:
		return 0,nil
	case *StringEntry:
		return len(entry.(*StringEntry).Value()),nil
	default:
		return 0,ErrWrongType
	}
}

func (cache *Cache) Append(key string, value string) (int, error) {
	entry := cache.get(key)
	switch entry.(type) {
	case nil:
		entry := new(StringEntry)
		entry.SetValue(value)
		cache.Database.SetString(key, entry)
	case *StringEntry:
		value += entry.(*StringEntry).Value()
		entry.(*StringEntry).SetValue(value)
	default:
		return 0, ErrWrongType
	}

	return len(value), nil
}

func (cache *Cache) SetRange(key string, pos int, replace string) (int, error) {
	if pos < 0{
		return 0, ErrOffsetOutOfRange
	}

	entry := cache.get(key)
	switch entry.(type) {
	case nil:
		str := make([]byte, pos)
		for i := 0; i< pos; i++{
			str[i] = '\x00'
		}
		entry := new(StringEntry)
		entry.SetValue(string(str) + replace)
		cache.Database.SetString(key, entry)

		return len(replace), nil
	case *StringEntry:
		str := []byte(entry.(*StringEntry).Value())
		for i:=0; i< len(replace); i++{
			p := pos + i
			if p < len(str){
				str[p] = replace[i]
			}else{
				str = append(str, replace[i])
			}
		}

		entry.(*StringEntry).SetValue(string(str))
		return len(str) , nil
	default:
		return 0, ErrWrongType
	}
}

func (cache *Cache) GetRange(key string, start, end int) (string, error) {
	entry := cache.get(key)
	switch entry.(type) {
	case nil:
		return "", nil
	case *StringEntry:
		val := entry.(*StringEntry).Value()
		if start < 0{
			start =  len(val) + start
		}
		if end < 0{
			entry = len(val) + end
		}

		var str []byte
		for i := start; i <= end; i++{
			if i >= len(val) {
				break
			}
			str = append(str, val[i])
		}

		return string(str), nil
	default:
		return "", ErrWrongType
	}
}

func (cache *Cache) Incr(key string) (int, error) {
	return cache.IncrBy(key, 1)
}

func (cache *Cache) IncrBy(key string, increment int) (int, error) {
	entry := cache.get(key)
	switch entry.(type) {
	case nil:
		entry := new(StringEntry)
		entry.SetValue(strconv.Itoa(increment))
		cache.Database.SetString(key, entry)
		return 1, nil
	case *StringEntry:
		if n, ok := entry.(*StringEntry).Incr(1);ok{
			return int(n), nil
		}else{
			return 0, ErrIntegerOrOutOfRange
		}
	default:
		return 0, ErrIntegerOrOutOfRange
	}
}

func (cache *Cache) IncrByFloat(key string, increment float64) (float64, error) {
	panic("implement me")
}

func (cache *Cache) Decr(key string) (int, error) {
	return cache.IncrBy(key, -1)
}

func (cache *Cache) DecrBy(key string, increment int) (int, error) {
	return cache.IncrBy(key, increment)
}

func (cache *Cache) MSet(kv map[string]string) string {
	for k, v := range kv{
		cache.Set(k, v)
	}

	return OK
}

func (cache *Cache) MSetNX(kv map[string]string) int {
	n := 0

	for k,v := range kv{
		n += cache.SetNX(k, v)
	}

	return n
}

func (cache *Cache) MGet(keys ...string) []interface{} {
	var values = make([]interface{}, len(keys))
	for id, k := range keys{
		v, err := cache.Get(k)
		if err != nil{
			values[id] = err
		}else{
			values[id] = v
		}
	}
	return values
}
