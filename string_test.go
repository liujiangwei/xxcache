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
	value := "this is a"
	key := "string_test"

	cache.Set(key, value)

	if v, err := cache.Get(key); err != nil {
		t.Fail()
	} else if value != v {
		t.Fail()
	}
}

func TestXXCache_SetEX(t *testing.T) {
	cache := NewXXCache()
	value := "TestXXCache_SetEX"
	key := "string_test"
	cache.SetEX(key, value, time.Millisecond*2)
	if v, err := cache.Get(key); err != nil {
		t.Fail()
	} else if v != value {
		t.Fail()
	}
	time.Sleep(time.Millisecond)
	if _, err := cache.Get(key); err != nil {
		t.Fail()
	}

	time.Sleep(time.Millisecond)
	if _, err := cache.Get(key); err != nil {
	} else {
		t.Fail()
	}

}

func BenchmarkXXCache_Set(b *testing.B) {
	cache := NewXXCache()
	value := "BenchmarkXXCache_Set BenchmarkXXCache_Set BenchmarkXXCache_Set BenchmarkXXCache_Set BenchmarkXXCache_Set"
	value += value
	value += value
	value += value
	value += value

	for i := 0; i < b.N; i++ {
		str := string(i)
		cache.Set(str, value)
	}
}

func BenchmarkXXCache_Get(b *testing.B) {
	cache := NewXXCache()
	value := "BenchmarkXXCache_Get"
	value += value
	value += value
	value += value
	value += value
	value += value
	value += value
	cache.Set("test_string", value)
	for i := 0; i < b.N; i++ {
		cache.Get("test_string")
	}
}

func BenchmarkXXCache_SetEX(b *testing.B) {
	cache := NewXXCache()
	for i := 0; i < b.N; i++ {
		key := string(i)
		cache.SetEX(key, "test_string", time.Second*60)
	}
}
