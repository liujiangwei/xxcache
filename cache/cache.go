package cache

import (
	"errors"
	"github.com/cornelk/hashmap"
	"sync"
	"time"
)

type Database struct {
	dataDict    hashmap.HashMap
	expiresDict hashmap.HashMap
	lock        sync.Mutex
}

func (db *Database) LPush(key string, values ...string) (int, error) {
	panic("implement me")
}

func (db *Database) LPushX(key string, values ...string) (int, error) {
	panic("implement me")
}

func (db *Database) RPush(key string, values ...string) (int, error) {
	panic("implement me")
}

func (db *Database) RPushX(key string, values ...string) (int, error) {
	panic("implement me")
}

func (db *Database) LPop(key string) (string, error) {
	panic("implement me")
}

func (db *Database) RPop(key string) (string, error) {
	panic("implement me")
}

func (db *Database) RPopLPush(keyFrom, keyDestination string) (string, error) {
	panic("implement me")
}

func (db *Database) LRem(key string, count int, value string) (int, error) {
	panic("implement me")
}

func (db *Database) LLen(key string) (int, error) {
	panic("implement me")
}

func (db *Database) LIndex(key string, index int) (string, error) {
	panic("implement me")
}

func (db *Database) LInsert(key string, direction string, pivot string, value string) (int, error) {
	panic("implement me")
}

func (db *Database) LSet(key string, index int, value string) (string, error) {
	panic("implement me")
}

func (db *Database) LRange(key string, start int, stop int) ([]string, error) {
	panic("implement me")
}

func (db *Database) LTrim(key string, start int, stop int) (string, error) {
	panic("implement me")
}

func (db *Database) BLPop(timeout time.Duration, keys ...string) {
	panic("implement me")
}

func (db *Database) BRPop(timeout time.Duration, keys ...string) {
	panic("implement me")
}

func (db *Database) BRPopLPush(keyFrom, keyDestination string, timeout time.Duration) {
	panic("implement me")
}

var ErrKeyNil = errors.New("redis nil")
var ErrWrongType = errors.New("wrong type,operation against a key holding the wrong kind of value")
var ErrOffsetOutOfRange = errors.New("offset is out of range")
var ErrIntegerOrOutOfRange = errors.New("value is not an integer or out of range")

var OK = "OK"
