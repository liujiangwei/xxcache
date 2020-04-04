package entry

import "github.com/liujiangwei/xxcache/dict"

type Database struct {
	Dict dict.Dict
}

func (db *Database) Get(key Key)(interface{}, bool){
	e, ok := db.Dict.Get(key)
	return e, ok
}

func (db *Database)Set(key Key, value interface{}){
	db.Dict.Set(key, value)
}