package cache

import (
	"strconv"
	"testing"
	"time"
)

var cache = New()

func init() {
	//cache.Cron()
}

func TestCache_Set(t *testing.T) {
	if ok, err := cache.Set("test", "test"); err != nil || ok != OK{
		t.Fatal(ok, err)
	}
	cache.Del("test")
}

func TestCache_SetNX(t *testing.T) {
	if ok, err := cache.Set("test", "test"); err != nil || ok != OK {
		t.Fail()
	}

	if n, err := cache.SetNX("test", "test"); err != nil || n == 1 {
		t.Fail()
	}

	cache.Del("test")
	if n, err := cache.SetNX("test", "test"); err != nil || n == 0 {
		t.Fatal(err, n)
	}

	cache.Del("test")
}

func TestCache_SetEX(t *testing.T) {
	cache.Del("test")

	if ok, err := cache.SetEX("test", "test", 1); err != nil || ok != OK{
		t.Fatal(ok, err)
	}

	if val, err := cache.Get("test"); err != nil || val != "test"{
		t.Fatal(val, err)
	}

	time.Sleep(time.Second)

	if ok, err := cache.Get("test"); err != ErrKeyNil || ok == OK{
		t.Fatal(ok, err)
	}

	cache.Del("test")
}
func TestCache_PSetEX(t *testing.T) {
	if _, err := cache.PSetEX("test", "test", 100); err != nil {
		t.Fail()
	}

	if _, err := cache.Get("test"); err != nil {
		t.Fail()
	}

	time.Sleep(time.Millisecond * 100)

	if _, err := cache.Get("test"); err == nil {
		t.Fail()
	}

	cache.Del("test")
}

func TestCache_Get(t *testing.T) {
	cache.Del("test")
	if ok, err := cache.Set("test", "test"); err != nil || ok != OK{
		t.Fatal(ok, err)
	}

	if val, err := cache.Get("test"); err != nil || val != "test" {
		t.Fatal(val, err)
	}

	// test for int
	if ok, err := cache.Set("test", "1"); err != nil || ok != OK{
		t.Fatal(ok, err)
	}
	if val, err := cache.Get("test"); err != nil || val != "1" {
		t.Fatal(val, err)
	}

	// test for float
	if ok, err := cache.Set("test", "1.11"); err != nil || ok != OK{
		t.Fatal(ok, err)
	}

	if val, err := cache.Get("test"); err != nil || val != "1.11" {
		t.Fatal(val, err)
	}

	cache.Del("test")
}

func TestCache_GetSet(t *testing.T) {
	cache.Del("test")

	// first get set should return nil error
	if _, err := cache.GetSet("test", "test"); err == nil {
		t.Fail()
	}

	if val, err := cache.GetSet("test", "test1"); err != nil || val != "test" {
		t.Fail()
	}

	cache.Del("test")
}

func TestCache_StrLen(t *testing.T) {
	cache.Del("test")

	if n, err := cache.StrLen("test"); err != nil || n != 0 {
		t.Fatal(n, err)
	}

	if ok, err := cache.Set("test", "test"); err != nil || ok != OK {
		t.Fatal(ok, err)
	}

	if n, err := cache.StrLen("test"); err != nil || n != 4 {
		t.Fatal(n, err)
	}

	cache.Del("test")
}

func TestCache_Append(t *testing.T) {
	cache.Del("test")
	if n, err := cache.Append("test", "t"); err != nil || n !=1 {
		t.Fatal(n, err)
	}

	if n, err := cache.Append("test", "t"); err != nil || n !=2 {
		t.Fatal(n, err)
	}
	cache.Del("test")

	// test for int
	if n, err := cache.Append("test", "1"); err != nil || n !=1 {
		t.Fatal(n, err)
	}

	if n, err := cache.Append("test", "1"); err != nil || n !=2 {
		t.Fatal(n, err)
	}
	cache.Del("test")

	// test for int
	if n, err := cache.Append("test", "1"); err != nil || n !=1 {
		t.Fatal(n, err)
	}

	if n, err := cache.Append("test", "a"); err != nil || n !=2 {
		t.Fatal(n, err)
	}
	cache.Del("test")

	// test for int
	if n, err := cache.Append("test", "0.1"); err != nil || n !=3 {
		t.Fatal(n, err)
	}

	if n, err := cache.Append("test", "a"); err != nil || n !=4 {
		t.Fatal(n, err)
	}

	cache.Del("test")
}

func TestCache_SetRange(t *testing.T) {
	cache.Del("test")
	if n , err := cache.SetRange("test", 0, "test");err != nil || n != 4{
		t.Fatal(n, err)
	}
	if val, err := cache.Get("test"); err != nil || val != "test"{
		t.Fatal(val, err)
	}

	cache.Del("test")
	if n , err := cache.SetRange("test", 1, "test");err != nil || n != 5{
		t.Fatal(n, err)
	}

	if val, err := cache.Get("test"); err != nil || val != "\x00test"{
		t.Fatal(val, err)
	}
	cache.Del("test")

	// set test = test
	if ok, err := cache.Set("test", "test"); err != nil || ok != OK{
		t.Fatal(ok, err)
	}
	// change to text
	if n, err := cache.SetRange("test", 2 , "x"); err !=nil || n !=4{
		t.Fatal(n, err)
	}
	if val, err := cache.Get("test"); err != nil || val != "text"{
		t.Fatal(val, err)
	}

	// change text to texts
	if n, err := cache.SetRange("test", 3 , "ts"); err !=nil || n !=5{
		t.Fatal(n, err)
	}
	if val, err := cache.Get("test"); err != nil || val != "texts"{
		t.Fatal(val, err)
	}
	cache.Del("test")

	// set test = test
	if ok, err := cache.Set("test", "t"); err != nil || ok != OK{
		t.Fatal(ok, err)
	}

	if n, err := cache.SetRange("test", 2, "st"); err != nil || n != 4{
		t.Fatal(n, err)
	}
	if val, err := cache.Get("test"); err != nil || val != "t\x00st"{
		t.Fatal(val, err)
	}

	cache.Del("test")
}

func TestCache_GetRange(t *testing.T) {
	if val, err := cache.GetRange("test", 0 , 2); err != nil || val != ""{
		t.Fatal(val, err)
	}

	// set test = test
	if ok, err := cache.Set("test", "test"); err != nil || ok != OK{
		t.Fatal(ok, err)
	}

	if val, err := cache.GetRange("test", 0, 1); err != nil || val != "te"{
		t.Fatal(val, err)
	}

	if val, err := cache.GetRange("test", 0, -1); err != nil || val != "test"{
		t.Fatal(val, err)
	}

	if val, err := cache.GetRange("test", 0, -2); err != nil || val != "tes"{
		t.Fatal(val,err)
	}

	if val, err := cache.GetRange("test", -1, -1); err != nil || val != "t"{
		t.Fatal(val, err)
	}

	if val, err := cache.GetRange("test", -2, -1); err != nil || val != "st"{
		t.Fatal(val, err)
	}
}
func TestCache_Incr(t *testing.T) {
	cache.Del("test")

	if n, err := cache.Incr("test"); err != nil || n != 1 {
		t.Fail()
	}

	if n, err := cache.Incr("test"); err != nil || n != 2 {
		t.Fail()
	}

	cache.Del("test")
}

func TestCache_IncrBy(t *testing.T) {
	cache.Del("test")

	if n, err := cache.IncrBy("test", 2); err != nil || n != 2 {
		t.Fail()
	}

	if n, err := cache.IncrBy("test", 1); err != nil || n != 3 {
		t.Fail()
	}

	if n, err := cache.IncrBy("test", -2); err != nil || n != 1 {
		t.Fail()
	}

	cache.Del("test")
}

func TestCache_IncrByFloat(t *testing.T) {
	cache.Del("test")

	if n, err := cache.IncrByFloat("test", 2.1); err != nil || n != 2.1 {
		t.Fail()
	}

	if n, err := cache.IncrByFloat("test", 1.1); err != nil || n != 3.2 {
		t.Fail()
	}

	if n, err := cache.IncrByFloat("test", 1); err != nil || n != 4.2 {
		t.Fail()
	}

	if n, err := cache.IncrByFloat("test", -1); err != nil || n != 3.2 {
		t.Fail()
	}

	if n, err := cache.IncrByFloat("test", -1.1); err != nil || n != 2.1 {
		t.Fail()
	}

	cache.Del("test")
}

func TestCache_Decr(t *testing.T) {
	cache.Del("test")

	if n, err := cache.Decr("test"); err != nil || n != -1 {
		t.Fail()
	}

	if n, err := cache.Decr("test"); err != nil || n != -2 {
		t.Fail()
	}

	cache.Del("test")
}

func TestCache_DecrBy(t *testing.T) {
	if n, err := cache.DecrBy("test", 1); err != nil || n != -1 {
		t.Fatal(n, err)
	}

	if n, err := cache.DecrBy("test", 1); err != nil || n != -2 {
		t.Fatal(n, err)
	}
}

func TestCache_MSet(t *testing.T) {
	var kv = map[string]string{
		"a" : "a",
		"b" : "b",
		"c" : "c",
		"d" : "d",
	}

	if ok := cache.MSet(kv); ok != OK{
		t.Fail()
	}

	for k,v := range kv{
		if val, err := cache.Get(k); err != nil || val != v{
			t.Fail()
		}
	}
}

func TestCache_MSetNX(t *testing.T) {
	var kv = map[string]string{
		"a" : "a",
		"b" : "b",
		"c" : "c",
		"d" : "d",
	}

	for k := range kv {
		cache.Del(k)
	}

	if n := cache.MSetNX(kv); n != len(kv){
		t.Fatal(n ,len(kv))
	}

	for k,v := range kv{
		if val, err := cache.Get(k); err != nil || val != v{
			t.Fatal(err, val)
		}
	}

	if n := cache.MSetNX(kv); n != 0{
		t.Fatal(n)
	}
}

func BenchmarkCache_Set(b *testing.B) {
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for i:=0; pb.Next(); i++{
			cache.Set(strconv.Itoa(i), "test")
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_Get(b *testing.B) {
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for i:=0; pb.Next(); i++{
			cache.Get(strconv.Itoa(i))
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_SetEX(b *testing.B) {
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for i:=0; pb.Next(); i++{
			cache.SetEX(strconv.Itoa(i), "test", 1)
		}
	})

	b.ReportAllocs()
}