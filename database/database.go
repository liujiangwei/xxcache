package database

import (
	"github.com/cornelk/hashmap"
	"time"
)
import "github.com/sean-public/fast-skiplist"

type Database struct {
	Dict    hashmap.HashMap
	Expires hashmap.HashMap
}

func (db *Database) Get(key string) *Entry {
	value, ok := db.Dict.Get(key)
	if !ok {
		return nil
	}

	if expires, ok := db.Expires.Get(key); ok{
		return value.(*Entry)
	}else if expiresMs, ok := expires.(int64); !ok{
		return nil
	}else if expiresMs < time.Now().UnixNano()/1000000{
		return nil
	}

	return value.(*Entry)
}

func (db *Database) Set(key string, entry Entry) {
	db.Dict.Set(key, entry)
}


type StringEntry struct {
	Val string
}


type ListEntry struct {
	Val []string
}


type ZSetEntry struct {
	Val *skiplist.SkipList
}

type HashEntry struct {
	Val *hashmap.HashMap
}

type SetEntry struct {
	Val *hashmap.HashMap
}

type Entry interface {
}


