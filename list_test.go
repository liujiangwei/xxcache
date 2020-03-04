package xxcache

import "testing"

func TestXXCache_RPush(t *testing.T) {
	cache := NewXXCache()
	cache.RPush("list_test", "1")

}

func TestXXCache_LPop(t *testing.T) {
	cache := NewXXCache()
	t.Log(cache.RPush("list_test", "1"))
	t.Log(cache.RPush("list_test", "2"))
	t.Log(cache.LPop("list_test"))
	t.Log(cache.LPop("list_test"))
	t.Log(cache.LPop("list_test"))
	t.Log(cache.RPush("list_test", "3"))
	t.Log(cache.RPush("list_test", "4"))
	t.Log(cache.LPop("list_test"))
	t.Log(cache.LPop("list_test"))
	t.Log(cache.LPop("list_test"))
}
