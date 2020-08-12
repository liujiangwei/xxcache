package cache

import (
	"testing"
	"time"
)

func TestStringEntry_Get(t *testing.T) {
	entry := StringEntry{data:"a"}
	if entry.Value() != "a"{
		t.Failed()
	}
}

func TestStringEntry_Set(t *testing.T) {
	entry := StringEntry{}
	entry.SetValue("a")
	if entry.Value() != "a"{
		t.Failed()
	}
}

func TestCache_Set(t *testing.T) {
	cache := new(Cache)

	if _,err := cache.Set("a", "a"); err != nil{
		t.Failed()
	}
}

func TestCache_Get(t *testing.T) {
	cache := new(Cache)

	cache.Set("a", "a")
	if val, err := cache.Get("a"); err != nil || val != "a"{
		t.Failed()
	}
}

func TestCache_SetNX(t *testing.T) {
	cache := new(Cache)
	cache.Set("a", "a")

	if n := cache.SetNX("a", "a"); n != 0{
		t.Failed()
	}

	if n := cache.SetNX("b", "b"); n != 1{
		t.Failed()
	}

	if n := cache.SetNX("b", "b"); n != 0{
		t.Failed()
	}
}

func TestCache_SetEX(t *testing.T) {
	cache := new(Cache)

	if _,err := cache.SetEX("a", "a",1); err != nil{
		t.Fatal(err)
	}

	if val, err := cache.Get("a"); err != nil{
		t.Fatal(err)
	}else if val != "a"{
		t.Fatal(err)
	}

	time.Sleep(time.Second)

	if _, err := cache.Get("a"); err != ErrKeyNil{
		t.Failed()
	}
}