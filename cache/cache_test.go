package cache_test

import (
	"github.com/liujiangwei/xxcache/cache"
	"strconv"
	"testing"
	"time"
)

var client = cache.New()

func init() {
	//client.Cron()
}

func TestCache_Set(t *testing.T) {
	if ok, err := client.Set("test", "test"); err != nil || ok != cache.OK{
		t.Fatal(ok, err)
	}
	client.Del("test")
}

func TestCache_SetNX(t *testing.T) {
	if ok, err := client.Set("test", "test"); err != nil || ok != cache.OK {
		t.Fail()
	}

	if n, err := client.SetNX("test", "test"); err != nil || n == 1 {
		t.Fail()
	}

	client.Del("test")
	if n, err := client.SetNX("test", "test"); err != nil || n == 0 {
		t.Fatal(err, n)
	}

	client.Del("test")
}

func TestCache_SetEX(t *testing.T) {
	client.Del("test")

	if ok, err := client.SetEX("test", "test", 1); err != nil || ok != cache.OK{
		t.Fatal(ok, err)
	}

	if val, err := client.Get("test"); err != nil || val != "test"{
		t.Fatal(val, err)
	}

	time.Sleep(time.Second)

	if ok, err := client.Get("test"); err != cache.ErrKeyNil || ok == cache.OK{
		t.Fatal(ok, err)
	}

	client.Del("test")
}
func TestCache_PSetEX(t *testing.T) {
	if _, err := client.PSetEX("test", "test", 100); err != nil {
		t.Fail()
	}

	if _, err := client.Get("test"); err != nil {
		t.Fail()
	}

	time.Sleep(time.Millisecond * 100)

	if _, err := client.Get("test"); err == nil {
		t.Fail()
	}

	client.Del("test")
}

func TestCache_Get(t *testing.T) {
	client.Del("test")
	if ok, err := client.Set("test", "test"); err != nil || ok != cache.OK{
		t.Fatal(ok, err)
	}

	if val, err := client.Get("test"); err != nil || val != "test" {
		t.Fatal(val, err)
	}

	// test for int
	if ok, err := client.Set("test", "1"); err != nil || ok != cache.OK{
		t.Fatal(ok, err)
	}
	if val, err := client.Get("test"); err != nil || val != "1" {
		t.Fatal(val, err)
	}

	// test for float
	if ok, err := client.Set("test", "1.11"); err != nil || ok != cache.OK{
		t.Fatal(ok, err)
	}

	if val, err := client.Get("test"); err != nil || val != "1.11" {
		t.Fatal(val, err)
	}

	client.Del("test")
}

func TestCache_GetSet(t *testing.T) {
	client.Del("test")

	// first get set should return nil error
	if _, err := client.GetSet("test", "test"); err == nil {
		t.Fail()
	}

	if val, err := client.GetSet("test", "test1"); err != nil || val != "test" {
		t.Fail()
	}

	client.Del("test")
}

func TestCache_StrLen(t *testing.T) {
	client.Del("test")

	if n, err := client.StrLen("test"); err != nil || n != 0 {
		t.Fatal(n, err)
	}

	if ok, err := client.Set("test", "test"); err != nil || ok != cache.OK {
		t.Fatal(ok, err)
	}

	if n, err := client.StrLen("test"); err != nil || n != 4 {
		t.Fatal(n, err)
	}

	client.Del("test")
}

func TestCache_Append(t *testing.T) {
	client.Del("test")
	if n, err := client.Append("test", "t"); err != nil || n !=1 {
		t.Fatal(n, err)
	}

	if n, err := client.Append("test", "t"); err != nil || n !=2 {
		t.Fatal(n, err)
	}
	client.Del("test")

	// test for int
	if n, err := client.Append("test", "1"); err != nil || n !=1 {
		t.Fatal(n, err)
	}

	if n, err := client.Append("test", "1"); err != nil || n !=2 {
		t.Fatal(n, err)
	}
	client.Del("test")

	// test for int
	if n, err := client.Append("test", "1"); err != nil || n !=1 {
		t.Fatal(n, err)
	}

	if n, err := client.Append("test", "a"); err != nil || n !=2 {
		t.Fatal(n, err)
	}
	client.Del("test")

	// test for int
	if n, err := client.Append("test", "0.1"); err != nil || n !=3 {
		t.Fatal(n, err)
	}

	if n, err := client.Append("test", "a"); err != nil || n !=4 {
		t.Fatal(n, err)
	}

	client.Del("test")
}

func TestCache_SetRange(t *testing.T) {
	client.Del("test")
	if n , err := client.SetRange("test", 0, "test");err != nil || n != 4{
		t.Fatal(n, err)
	}
	if val, err := client.Get("test"); err != nil || val != "test"{
		t.Fatal(val, err)
	}

	client.Del("test")
	if n , err := client.SetRange("test", 1, "test");err != nil || n != 5{
		t.Fatal(n, err)
	}

	if val, err := client.Get("test"); err != nil || val != "\x00test"{
		t.Fatal(val, err)
	}
	client.Del("test")

	// set test = test
	if ok, err := client.Set("test", "test"); err != nil || ok != cache.OK{
		t.Fatal(ok, err)
	}
	// change to text
	if n, err := client.SetRange("test", 2 , "x"); err !=nil || n !=4{
		t.Fatal(n, err)
	}
	if val, err := client.Get("test"); err != nil || val != "text"{
		t.Fatal(val, err)
	}

	// change text to texts
	if n, err := client.SetRange("test", 3 , "ts"); err !=nil || n !=5{
		t.Fatal(n, err)
	}
	if val, err := client.Get("test"); err != nil || val != "texts"{
		t.Fatal(val, err)
	}
	client.Del("test")

	// set test = test
	if ok, err := client.Set("test", "t"); err != nil || ok != cache.OK{
		t.Fatal(ok, err)
	}

	if n, err := client.SetRange("test", 2, "st"); err != nil || n != 4{
		t.Fatal(n, err)
	}
	if val, err := client.Get("test"); err != nil || val != "t\x00st"{
		t.Fatal(val, err)
	}

	client.Del("test")
}

func TestCache_GetRange(t *testing.T) {
	if val, err := client.GetRange("test", 0 , 2); err != nil || val != ""{
		t.Fatal(val, err)
	}

	// set test = test
	if ok, err := client.Set("test", "test"); err != nil || ok != cache.OK{
		t.Fatal(ok, err)
	}

	if val, err := client.GetRange("test", 0, 1); err != nil || val != "te"{
		t.Fatal(val, err)
	}

	if val, err := client.GetRange("test", 0, -1); err != nil || val != "test"{
		t.Fatal(val, err)
	}

	if val, err := client.GetRange("test", 0, -2); err != nil || val != "tes"{
		t.Fatal(val,err)
	}

	if val, err := client.GetRange("test", -1, -1); err != nil || val != "t"{
		t.Fatal(val, err)
	}

	if val, err := client.GetRange("test", -2, -1); err != nil || val != "st"{
		t.Fatal(val, err)
	}
}
func TestCache_Incr(t *testing.T) {
	client.Del("test")

	if n, err := client.Incr("test"); err != nil || n != 1 {
		t.Fail()
	}

	if n, err := client.Incr("test"); err != nil || n != 2 {
		t.Fail()
	}

	client.Del("test")
}

func TestCache_IncrBy(t *testing.T) {
	client.Del("test")

	if n, err := client.IncrBy("test", 2); err != nil || n != 2 {
		t.Fail()
	}

	if n, err := client.IncrBy("test", 1); err != nil || n != 3 {
		t.Fail()
	}

	if n, err := client.IncrBy("test", -2); err != nil || n != 1 {
		t.Fail()
	}

	client.Del("test")
}

func TestCache_IncrByFloat(t *testing.T) {
	client.Del("test")

	if n, err := client.IncrByFloat("test", 2.1); err != nil || n != 2.1 {
		t.Fail()
	}

	if n, err := client.IncrByFloat("test", 1.1); err != nil || n != 3.2 {
		t.Fail()
	}

	if n, err := client.IncrByFloat("test", 1); err != nil || n != 4.2 {
		t.Fail()
	}

	if n, err := client.IncrByFloat("test", -1); err != nil || n != 3.2 {
		t.Fail()
	}

	if n, err := client.IncrByFloat("test", -1.1); err != nil || n != 2.1 {
		t.Fail()
	}

	client.Del("test")
}

func TestCache_Decr(t *testing.T) {
	client.Del("test")

	if n, err := client.Decr("test"); err != nil || n != -1 {
		t.Fail()
	}

	if n, err := client.Decr("test"); err != nil || n != -2 {
		t.Fail()
	}

	client.Del("test")
}

func TestCache_DecrBy(t *testing.T) {
	if n, err := client.DecrBy("test", 1); err != nil || n != -1 {
		t.Fatal(n, err)
	}

	if n, err := client.DecrBy("test", 1); err != nil || n != -2 {
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

	if ok := client.MSet(kv); ok != cache.OK{
		t.Fail()
	}

	for k,v := range kv{
		if val, err := client.Get(k); err != nil || val != v{
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
		client.Del(k)
	}

	if n := client.MSetNX(kv); n != len(kv){
		t.Fatal(n ,len(kv))
	}

	for k,v := range kv{
		if val, err := client.Get(k); err != nil || val != v{
			t.Fatal(err, val)
		}
	}

	if n := client.MSetNX(kv); n != 0{
		t.Fatal(n)
	}
}

func BenchmarkCache_Set(b *testing.B) {
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for i:=0; pb.Next(); i++{
			if ok, err := client.Set(strconv.Itoa(i), "test"); err != nil || ok != cache.OK{
				b.Fatal(ok, err)
			}
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_Get(b *testing.B) {
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for i:=0; pb.Next(); i++{
			if val, err := client.Get(strconv.Itoa(i)); err != nil{
				b.Fatal(val, err)
			}
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_SetEX(b *testing.B) {
	b.SetParallelism(4)
	b.RunParallel(func(pb *testing.PB) {
		for i:=0; pb.Next(); i++{
			if ok, err := client.SetEX(strconv.Itoa(i), "test", 1); err != nil || ok != cache.OK{
				b.Fatal(ok, err)
			}
		}
	})

	b.ReportAllocs()
}