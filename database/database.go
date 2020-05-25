package database

import (
	"github.com/liujiangwei/xxcache/dict"
	"time"
)

type Database struct {
	dict      dict.Dict
	expires   dict.Dict
	keyNumber int
	expireId  int
}

func NewDatabase(d dict.Dict) *Database {
	return &Database{
		dict:    d,
		expires: dict.Default(),
	}
}

func (db *Database) Get(key string) (Entry, bool) {
	v, ok := db.dict.Get(key)

	if ok && db.keyIsExpired(key) {
		ok = false
		db.Delete(key)
	}

	return v, ok
}

func (db *Database) SetWithExpire(key string, value interface{}, expires time.Duration) {
	db.Set(key, value)
	db.Expired(key, expires)
}

func (db *Database) Set(key Key, value interface{}) {
	db.dict.Set(key, value)
}

func (db *Database) Delete(key Key) {
	db.dict.Del(key)
	db.expires.Del(key)
}

func (db *Database) Expired(key string, expires time.Duration) {
	if expires <= 0 {
		db.expires.Del(key)
		db.dict.Del(key)

		return
	}

	db.expires.Set(key, time.Now().Add(expires))
}

func (db *Database) keyIsExpired(key string) bool {
	expire, ok := db.expires.Get(key)
	if !ok {
		return false
	}

	if expire, ok := expire.(time.Time); ok {
		if expire.After(time.Now()) {
			return false
		} else {
			db.expires.Del(key)
			return true
		}
	} else {
		db.expires.Del(key)
		return false
	}
}
