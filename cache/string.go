package cache

import (
	"errors"
	"strconv"
	"time"
)

func (c *Cache) Set(key, value string) (string, error) {
	c.dataDict.Set(key, tryInt64(value))
	return OK, nil
}

func tryInt64(value string) interface{} {
	if n, err := strconv.Atoi(value); err == nil {
		// try convert string value to int64
		return n
	} else if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	return value
}

// SET if Not eXists
func (c *Cache) SetNX(key, value string) (int, error) {
	if _, loaded := c.dataDict.GetOrInsert(key, tryInt64(value)); loaded {
		return 1, nil
	} else {
		return 0, nil
	}
}

// SETEX key seconds value
func (c *Cache) SetEX(key, value string, expires int64) (string, error) {
	c.dataDict.Set(key, tryInt64(value))
	c.expiresDict.Set(key, time.Second*time.Duration(expires))

	return OK, nil
}

func (c *Cache) PSetEX(key, value string, expires int64) (string, error) {
	c.dataDict.Set(key, tryInt64(value))
	c.expiresDict.Set(key, time.Millisecond*time.Duration(expires))

	return OK, nil
}

func fmtString(v interface{}) (str string, err error) {
	switch value := v.(type) {
	case int:
		str = strconv.Itoa(value)
	case int64:
		str = strconv.Itoa(int(value))
	default:
		err = ErrWrongType
	}

	return str, err
}

// if key is expired
func (c *Cache) expired(key string) bool {
	v, ok := c.expiresDict.Get(key)
	if !ok || v == nil {
		return false
	}

	if expires, ok := v.(time.Time); ok {
		return expires.Before(time.Now())
	}

	return false
}

// set expire time after duration from now
func (c *Cache) expires(key string, duration time.Duration) {
	c.expiresDict.Set(key, time.Now().Add(duration))
}

func (c *Cache) Get(key string) (string, error) {
	val, ok := c.dataDict.Get(key)
	if !ok || c.expired(key) {
		return "", ErrKeyNil
	}

	return fmtString(val)
}

func (c *Cache) GetSet(key, value string) (string, error) {
	old, err := c.Get(key)
	c.dataDict.Set(key, tryInt64(value))
	return old, err
}

func (c *Cache) StrLen(key string) (int, error) {
	val, ok := c.dataDict.Get(key)
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
	old, ok := c.dataDict.Get(key)
	if !ok || c.expired(key) {
		c.dataDict.Set(key, tryInt64(value))
		return len(value), nil
	}

	if old, err := fmtString(old); err != nil {
		return 0, err
	} else {
		value = old + value
	}

	c.dataDict.Set(key, tryInt64(value))
	return len(value), nil
}

func (c *Cache) SetRange(key string, pos int, replace string) (int, error) {
	if pos < 0 {
		return 0, ErrOffsetOutOfRange
	}

	// key is not exist or expired
	old, ok := c.dataDict.Get(key)
	if !ok || c.expired(key) {
		str := make([]byte, pos)
		for i := 0; i < pos; i++ {
			str[i] = '\x00'
		}

		replace = string(str) + replace
		c.dataDict.Set(key, tryInt64(replace))
		return len(replace), nil
	}

	oldStr, err := fmtString(old)
	// not string value
	if err != nil {
		return 0, err
	}
	str := []byte(oldStr)
	for i := 0; i < len(replace); i++ {
		p := pos + i
		if p < len(str) {
			str[p] = replace[i]
		} else {
			str = append(str, replace[i])
		}
	}

	c.dataDict.Set(key, tryInt64(string(str)))
	return len(str), nil
}

func (c *Cache) GetRange(key string, start, end int) (string, error) {
	old, ok := c.dataDict.Get(key)
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

var ErrIncrChanged = errors.New("value is reset")

func (c *Cache) IncrBy(key string, increment int64) (int, error) {
	actual, ok := c.dataDict.GetOrInsert(key, increment)
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

	if c.dataDict.Cas(key, actual, val) {
		return n, nil
	} else {
		return 0, ErrIncrChanged
	}
}

func (c *Cache) IncrByFloat(key string, increment float64) (float64, error) {
	actual, ok := c.dataDict.GetOrInsert(key, increment)
	if !ok {
		return increment, nil
	}

	switch val := actual.(type) {
	case float64:
		val += increment
		if c.dataDict.Cas(key, actual, val) {
			return val, nil
		} else {
			return 0, ErrIncrChanged
		}
	case int:
		f := float64(val) + increment
		if c.dataDict.Cas(key, actual, f) {
			return f, nil
		} else {
			return 0, ErrIncrChanged
		}
	default:
		return 0, ErrWrongType
	}
}

func (c *Cache) Decr(key string) (int, error) {
	return c.IncrBy(key, -1)
}

func (c *Cache) DecrBy(key string, increment int64) (int, error) {
	return c.IncrBy(key, increment)
}

func (c *Cache) MSet(kv map[string]string) string {
	for k, v := range kv {
		c.dataDict.Set(k, tryInt64(v))
	}

	return OK
}

func (c *Cache) MSetNX(kv map[string]string) int {
	n := 0

	for k, v := range kv {
		if _, loaded := c.dataDict.GetOrInsert(k, tryInt64(v)); loaded {
			n++
		}
	}

	return n
}

func (c *Cache) MGet(keys ...string) []interface{} {
	var values = make([]interface{}, len(keys))
	for id, k := range keys {
		v, err := c.Get(k)
		if err != nil {
			values[id] = err
		} else {
			values[id] = v
		}
	}
	return values
}
