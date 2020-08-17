package cache

import (
	"errors"
	"strconv"
)

type hashEntry struct {
	SyncMapDatabase
	length int
}

func (he *hashEntry) get(field string) (value string, ok bool) {
	v, ok := he.Load(field)
	if !ok {
		return "", ok
	}

	// here is must string
	value, _ = fmtString(v)
	return value, ok
}

func (he *hashEntry) set(field string, value string) (n int) {
	if _, loaded := he.LoadOrStore(field, tryInt64(value)); !loaded {
		n = 1
	} else {
		he.Store(field, tryInt64(value))
	}

	he.length += n

	return n
}

func newHash() *hashEntry {
	return &hashEntry{}
}

func (c *Cache) getHashEntry(key string) (*hashEntry, error) {
	entry, ok := c.get(key)
	if !ok {
		return nil, nil
	}

	if he, ok := entry.(*hashEntry); ok {
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

	return he.set(field, value), nil
}

//HGET
func (c *Cache) HSetNX(key string, field string, value string) (n int, err error) {
	he, err := c.getHashEntry(key)
	if err != nil {
		return 0, err
	}

	if he == nil {
		he = newHash()
		c.dataDict.Store(key, he)
	}

	if _, ok := he.LoadOrStore(field, value); !ok {
		he.length++
		n = 1
	}

	return n, nil
}

//HEXISTS
func (c *Cache) HExists(key string, field string) (n int, err error) {
	he, err := c.getHashEntry(key)
	if err != nil {
		return 0, err
	}

	if he == nil {
		return 0, nil
	}

	if _, ok := he.get(field); ok {
		n = 1
	}

	he.length -= n

	return n, err
}

//HDEL
func (c *Cache) HDel(key string, fields ...string) (n int, err error) {
	he, err := c.getHashEntry(key)
	if err != nil {
		return 0, err
	}

	if he == nil {
		return 0, nil
	}

	for _, field := range fields {
		if _, ok := he.get(field); ok {
			n++
			he.Delete(field)
		}
	}

	if n != 0{
		he.length -= n
	}

	return n, nil
}

//HLEN
func (c *Cache) HLen(key string) (n int, err error) {
	he, err := c.getHashEntry(key)
	if err != nil {
		return 0, err
	}

	if he == nil {
		return 0, nil
	}

	return he.length, nil
}

//HSTRLEN
func (c *Cache) HStrLen(key, field string) (n int, err error) {
	he, err := c.getHashEntry(key)
	if err != nil || he == nil {
		return n, err
	}

	if val, ok := he.get(field); ok {
		n = len(val)
	}

	return n, err
}

func (c *Cache) HIncrBy(key, field string, increment int) (n int, err error){
	he, err := c.getHashEntry(key)
	if err != nil{
		return n, err
	}

	if he == nil{
		he = newHash()
		c.dataDict.Store(key, he)
	}

	v, ok := he.Load(field)
	if !ok{
		he.set(field, strconv.Itoa(increment))
		return increment, nil
	}

	switch val := v.(type) {
	case string:
		if val == ""{
			n = increment
		}else{
			err = ErrHashValueIsNotFloat
		}
	case int:
		n = val + increment
	case int32:
		n = int(val) + increment
	case int64:
		n = int(val) + increment
	case uint32:
		n = int(val) + increment
	case uint64:
		n = int(val) + increment
	default:
		err = ErrHashValueIsNotFloat
	}

	if err == nil{
		he.set(field, strconv.Itoa(increment))
	}

	return n, err
}

var ErrHashValueIsNotFloat = errors.New("hash value is not a float")
//HINCRBY
func (c *Cache) HIncrByFloat(key, field string, increment float64) (f float64, err error){
	he, err := c.getHashEntry(key)
	if err != nil{
		return f, err
	}

	if he == nil{
		he = newHash()
		c.dataDict.Store(key, he)
	}

	v, ok := he.Load(field)
	if !ok{
		he.set(field, strconv.FormatFloat(increment, 'f', -1, 64))
		return increment, nil
	}

	switch val := v.(type) {
	case string:
		if val == ""{
			f = increment
		}else{
			err = ErrHashValueIsNotFloat
		}
	case int:
		f = float64(val) + increment
	case int32:
		f = float64(val) + increment
	case int64:
		f = float64(val) + increment
	case uint32:
		f = float64(val) + increment
	case uint64:
		f = float64(val) + increment
	default:
		err = ErrHashValueIsNotFloat
	}

	if err == nil{
		he.set(field, strconv.FormatFloat(increment, 'f', -1, 64))
	}

	return f, err
}
//HMSET
func (c *Cache) HMSet(key string, fv map[string]string) (n int, err error){
	he, err := c.getHashEntry(key)
	if err != nil{
		return n, err
	}

	if he == nil{
		he = newHash()
		c.dataDict.Store(key, he)
	}

	for f, v := range fv{
		n += he.set(f, v)
	}

	return n, err
}

//HMGET
func (c *Cache)HMGet(key string, fields ...string) ([]interface{}, error) {
	he, err := c.getHashEntry(key)
	if err != nil{
		return nil, err
	}

	var values = make([]interface{}, len(fields))
	for id, field := range fields{
		if he == nil{
			values[id] = ErrKeyNil
			continue
		}

		if value, ok := he.get(field); ok{
			values[id] = value
		}else{
			values[id] = ErrKeyNil
		}
	}

	return values, nil
}

//HKEYS
func(c *Cache) HKeys(key string)(fields []string,err error){
	he, err := c.getHashEntry(key)
	if err != nil{
		return nil, err
	}

	 he.Range(func(key, value interface{}) bool {
		if field, ok := key.(string); ok{
			fields = append(fields, field)
		}
	 	return true
	 })

	return fields, nil
}

//HVALS
func(c *Cache) HVals(key string)(values []string,err error){
	he, err := c.getHashEntry(key)
	if err != nil{
		return nil, err
	}

	he.Range(func(key, value interface{}) bool {
		if field, ok := value.(string); ok{
			values = append(values, field)
		}
		return true
	})

	return values, nil
}

//HGETALL
func(c *Cache) HGetAll(key string)(kv map[string]string,err error){
	he, err := c.getHashEntry(key)
	if err != nil{
		return nil, err
	}

	he.Range(func(key, value interface{}) bool {
		if field, ok := key.(string); ok{
			if v, err := fmtString(value); err != nil{
				kv[field] = v
			}
		}

		return true
	})

	return kv, nil
}
//HSCAN
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
	if str, ok := he.get(field); !ok {
		err = ErrKeyNil
	} else {
		value = str
	}

	return value, err
}
