package cache

import (
	"strconv"
	"time"
)

func (c *Cache) Set(key, value string) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.dataDict.Store(key, tryInt64(value))
	return OK, nil
}

func tryInt64(value string) interface{} {
	if n, err := strconv.Atoi(value); err == nil {
		// try convert string value to int
		return n
	} else if f, err := strconv.ParseFloat(value, 64); err == nil {
		// try float
		return f
	}

	return value
}

// SET if Not eXists
func (c *Cache) SetNX(key, value string) (int, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, loaded := c.dataDict.LoadOrStore(key, tryInt64(value)); loaded {
		return 0, nil
	} else {
		return 1, nil
	}
}

// SETEX key seconds value
func (c *Cache) SetEX(key, value string, expires int64) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.dataDict.Store(key, tryInt64(value))
	c.expires(key, time.Second*time.Duration(expires))

	return OK, nil
}

func (c *Cache) PSetEX(key, value string, expiresMs int64) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.dataDict.Store(key, tryInt64(value))
	c.expires(key, time.Millisecond*time.Duration(expiresMs))

	return OK, nil
}

func fmtString(v interface{}) (str string, err error) {
	switch value := v.(type) {
	case int:
		str = strconv.Itoa(value)
	case int64:
		str = strconv.Itoa(int(value))
	case float64:
		str = strconv.FormatFloat(value, 'f', -1, 64)
	case string:
		str = value
	default:
		err = ErrWrongType
	}

	return str, err
}

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

func (c *Cache) Get(key string) (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	val, ok := c.get(key)
	if !ok{
		return "", ErrKeyNil
	}

	return fmtString(val)
}

func (c *Cache) GetSet(key, value string) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	old, ok := c.get(key)

	c.dataDict.Store(key, tryInt64(value))

	if !ok{
		return "", ErrKeyNil
	}

	return fmtString(old)
}

func (c *Cache) StrLen(key string) (int, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	val, ok := c.dataDict.Load(key)
	if !ok || c.expired(key) {
		return 0, nil
	}

	if val, err := fmtString(val); err != nil {
		return 0, err
	} else {
		return len(val), nil
	}
}

func (c *Cache) Append(key string, value string) (int, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	old, ok := c.dataDict.Load(key)
	if !ok || c.expired(key) {
		c.dataDict.Store(key, tryInt64(value))
		return len(value), nil
	}

	if old, err := fmtString(old); err != nil {
		return 0, err
	} else {
		value = old + value
	}

	c.dataDict.Store(key, tryInt64(value))
	return len(value), nil
}

func (c *Cache) SetRange(key string, pos int, replace string) (int, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if pos < 0 {
		return 0, ErrOffsetOutOfRange
	}

	// key is not exist or expired
	old, ok := c.dataDict.Load(key)
	if !ok || c.expired(key) {
		str := make([]byte, pos)
		for i := 0; i < pos; i++ {
			str[i] = '\x00'
		}

		replace = string(str) + replace
		c.dataDict.Store(key, tryInt64(replace))
		return len(replace), nil
	}

	oldStr, err := fmtString(old)
	// not string value
	if err != nil {
		return 0, err
	}
	str := []byte(oldStr)
	for i := len(str); i < pos; i++{
		str = append(str, '\x00')
	}

	for i := 0; i < len(replace); i++ {
		p := pos + i
		if p < len(str) {
			str[p] = replace[i]
		} else {
			str = append(str, replace[i])
		}
	}

	c.dataDict.Store(key, tryInt64(string(str)))
	return len(str), nil
}

func (c *Cache) GetRange(key string, start, end int) (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	old, ok := c.dataDict.Load(key)
	if !ok || c.expired(key) {
		return "", nil
	}
	oldStr, err := fmtString(old)
	if err != nil {
		return "", err
	}

	if start < 0 {
		start = len(oldStr) + start
	}
	if end < 0 {
		end = len(oldStr) + end
	}

	var str []byte
	for i := start; i <= end; i++ {
		if i < 0 {
			continue
		}

		if i >= len(oldStr) {
			break
		}

		str = append(str, oldStr[i])
	}
	return string(str), nil
}

func (c *Cache) Incr(key string) (int, error) {
	return c.IncrBy(key, 1)
}

func (c *Cache) IncrBy(key string, increment int64) (int, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.incrBy(key, increment)
}

func (c *Cache) incrBy(key string, increment int64) (int, error) {
	actual, ok := c.dataDict.LoadOrStore(key, increment)
	if !ok {
		return int(increment), nil
	}

	var n int
	var val interface{}
	switch t := actual.(type) {
	case uint64:
		val = t + uint64(increment)
		n = int(t + uint64(increment))
	case uint32:
		val = t + uint32(increment)
		n = int(t + uint32(increment))
	case int64:
		val = t + increment
		n = int(t + increment)
	case int32:
		val = t + int32(increment)
		n = int(t + int32(increment))
	case int:
		val = t + int(increment)
		n = t + int(increment)
	default:
		return 0, ErrWrongType
	}

	c.dataDict.Store(key, val)

	return n,nil
}

func (c *Cache) IncrByFloat(key string, increment float64) (float64, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	actual, ok := c.dataDict.LoadOrStore(key, increment)
	if !ok {
		return increment, nil
	}

	switch val := actual.(type) {
	case float64:
		val += increment
		c.dataDict.Store(key, val)
		return val, nil
	case int:
		f := float64(val) + increment
		c.dataDict.Store(key, f)
		return f, nil
	default:
		return 0, ErrWrongType
	}
}

func (c *Cache) Decr(key string) (int, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.incrBy(key, -1)
}

func (c *Cache) DecrBy(key string, decrement int64) (int, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.incrBy(key, -decrement)

}

func (c *Cache) MSet(kv map[string]string) (ok string) {
	//c.lock.Lock()
	//defer c.lock.Unlock()

	for k, v := range kv {
		c.dataDict.Store(k, tryInt64(v))
	}

	return OK
}

func (c *Cache) MSetNX(kv map[string]string) int {
	//c.lock.Lock()
	//defer c.lock.Unlock()

	n := 0

	for k, v := range kv {
		if _, loaded := c.dataDict.LoadOrStore(k, tryInt64(v)); !loaded {
			n++
		}
	}

	return n
}

func (c *Cache) MGet(keys ...string) []interface{} {
	//c.lock.RLock()
	//defer c.lock.RUnlock()

	var values = make([]interface{}, len(keys))
	for id, k := range keys {

		if v, ok := c.get(k); !ok {
			values[id] = ErrKeyNil
		} else {
			if str, err := fmtString(v);err != nil{
				values[id] =  err
			}else{
				values[id] = str
			}
		}
	}
	return values
}
