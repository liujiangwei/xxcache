package xxcache

import (
	"testing"
	"time"
)

func TestXXCache_Set(t *testing.T) {
	cache := NewXXCache()
	cache.Set("string_test", "this is a")
}

func TestXXCache_Get(t *testing.T) {
	cache := NewXXCache()
	cache.Set("string_test", "this is a")

	if v, err := cache.Get("a"); err != nil {
		t.Fail()
	} else {
		t.Log("string_test", "=>", v)
	}
}

func TestXXCache_SetEX(t *testing.T) {
	cache := NewXXCache()
	t.Log(cache.SetEX("string_test", "a", time.Second * 3))

	t.Log(cache.Get("string_test"))

	time.Sleep(time.Second * 5)

	t.Log(cache.Get("string_test"))
}
