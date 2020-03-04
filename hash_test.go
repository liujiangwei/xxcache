package xxcache

import "testing"

func TestXXCache_HSet(t *testing.T) {
	cache := NewXXCache()
	key := "hash_test"
	if nil != cache.HSet(key, "field", "value"){
		t.Fail()
	}
}

func TestXXCache_HGet(t *testing.T) {
	cache := NewXXCache()
	key := "hash_test"
	field := "hash_field"
	value := "hash_field_value"
	if nil != cache.HSet(key, field, value){
		t.Fail()
	}

	if v, err := cache.HGet(key, field); err != nil{
		t.Fail()
	}else if v != value{
		t.Fail()
	}
}

func TestXXCache_HDel(t *testing.T) {
	cache := NewXXCache()
	key := "hash_test"
	field := "hash_field"
	value := "hash_field_value"
	if nil != cache.HSet(key, field, value){
		t.Fail()
	}

	if v, err := cache.HGet(key, field); err != nil{
		t.Fail()
	}else if v != value{
		t.Fail()
	}

	if nil != cache.HDel(key, field){
		t.Fail()
	}

	if _, err := cache.HGet(key, field); err == nil{
		t.Fail()
	}
}

func TestXXCache_HSetNX(t *testing.T) {
	cache := NewXXCache()
	key := "hash_test"
	field := "hash_field"
	value := "hash_field_value"
	if nil != cache.HSetNX(key, field, value){
		t.Fail()
	}

	if nil == cache.HSetNX(key, field, value){
		t.Fail()
	}
}