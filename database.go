package xxcache

import (
	"github.com/cornelk/hashmap"
	"time"
)
import "github.com/sean-public/fast-skiplist"

type Database struct {
	dict hashmap.HashMap
	expires hashmap.HashMap
}

func (db *Database) Get(key string) *Entry {
	value, ok := db.dict.Get(key)
	if !ok {
		return nil
	}

	if expires, ok := db.expires.Get(key); ok{
		return value.(*Entry)
	}else if expiresMs, ok := expires.(int64); !ok{
		return nil
	}else if expiresMs < time.Now().UnixNano()/1000000{
		return nil
	}

	return value.(*Entry)
}

func (db *Database) Set(key string, entry Entry) {
	db.dict.Set(key, entry)
}


type StringEntry struct {
	val string
}


type ListEntry struct {
	val []string
}


type ZSetEntry struct {
	val *skiplist.SkipList
}

type HashEntry struct {
	val *hashmap.HashMap
}

type SetEntry struct {
	val *hashmap.HashMap
}

type Entry interface {
}


