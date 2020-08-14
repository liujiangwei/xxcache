package cache_test

import "testing"

func TestCache_HSet(t *testing.T) {
	client.Del("test")
	if n, err := client.HSet("test","a", "b");err != nil || n !=1 {
		t.Fatal(n, err)
	}

	if n, err := client.HSet("test","a", "b");err != nil || n !=0 {
		t.Fatal(n, err)
	}
}

func TestCache_HGet(t *testing.T) {
	client.Del("test")
	if n, err := client.HSet("test","a", "b");err != nil || n !=1 {
		t.Fatal(n, err)
	}

	if val, err := client.HGet("test", "a"); err != nil || val != "b"{
		t.Fatal(val, err)
	}

	if val, err := client.HGet("test", "b"); err == nil{
		t.Fatal(val, err)
	}

	if val, err := client.HGet("test1", "b"); err == nil{
		t.Fatal(val, err)
	}
}