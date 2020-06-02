package xxcache

import "github.com/cornelk/hashmap"

type Database struct {
	dict hashmap.HashMap
	expires hashmap.HashMap
}

func (db *Database) Get(key string) *Entry {
	value, ok := db.dict.Get(key)
	if !ok {
		return nil
	}

	return value.(*Entry)
}

func (db *Database) Set(key string, entry *Entry) {
	db.dict.Set(key, entry)
}

type T uint

const (
	TString T = iota
	TList
	TSet
	TZSet
	THash
)

type Entry struct {
	t    T
	data interface{}
}
