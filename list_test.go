package xxcache

import "testing"

func TestXXCache_RPush(t *testing.T) {
	cache := NewXXCache()
	if err := cache.RPush("list_test", "1"); err != nil {
		t.Fail()
	}
}

func TestXXCache_LPop(t *testing.T) {
	cache := NewXXCache()
	var key = "list_test"

	for i := 0; i < 10; i++ {
		cache.RPush(key, string(i))
	}

	for i := 0; i < 10; i++ {
		if v, err := cache.LPop(key); err != nil {
			t.Fail()
		} else if v != string(i) {
			t.Fail()
		}
	}
}

func BenchmarkXXCache_RPush(b *testing.B) {
	key := "test_list"
	value := "BenchmarkXXCache_RPush"
	cache := NewXXCache()
	for i := 0; i < b.N; i++ {
		cache.RPush(key, value)
	}
}

func BenchmarkXXCache_LPop(b *testing.B) {
	key := "test_list"
	value := "BenchmarkXXCache_RPush"
	cache := NewXXCache()
	for i := 0; i < b.N; i++ {
		cache.LPop(key)
		cache.RPush(key, value)
	}
}
