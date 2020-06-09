package database

import (
	"github.com/cornelk/hashmap"
	"time"
)

import "github.com/sean-public/fast-skiplist"

type Database struct {
	dataDict    hashmap.HashMap
	expiresDict hashmap.HashMap
}

func (db *Database) expires(key string) bool{
	expires, ok := db.expiresDict.Get(key)
	if !ok {
		return false
	}

	if ms, ok := expires.(int64); ok{
		if ms == 0{
			return false
		}

		if ms * 1000 > time.Now().UnixNano(){
			return false
		}else{
			db.expiresDict.Del(key)
			return true
		}
	}else{
		return  false
	}
}

func (db *Database) Get(key string) Entry {
	value, ok := db.dataDict.Get(key)
	if !ok {
		return nil
	}

	if db.expires(key){
		db.dataDict.Del(key)
		return nil
	}

	return value.(Entry)
}

func (db *Database) set(key string, entry Entry) {
	if entry != nil{
		db.dataDict.Set(key, entry)
	}
}

func (db *Database) SetString(key string, entry *StringEntry){
	db.set(key, entry)
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


