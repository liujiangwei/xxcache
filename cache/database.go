package cache

import "sync"

type Database interface {
	Delete(key interface{})
	Store(key interface{}, value interface{})
 	Load(key interface{})(value interface{},ok bool)
	LoadOrStore(key interface{}, value interface{})(actual interface{}, loaded bool)
	Range(func(key interface{}, value interface{}) bool)
}

type SyncMapDatabase struct {
	sync.Map
}

