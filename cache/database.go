package cache

import (
	"github.com/cornelk/hashmap"
	"time"
)

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

func (db *Database) Set(key string, entry Entry) {
	if entry != nil{
		db.dataDict.Set(key, entry)
	}
}

// string
func (db *Database) SetString(key string, entry *StringEntry){
	db.Set(key, entry)
}

// list
func (db *Database) SetList(key string, entry *ListEntry){
	db.Set(key, entry)
}

type Entry interface {
}


