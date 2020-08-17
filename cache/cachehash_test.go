package cache_test

import "testing"

func TestCache_HSet(t *testing.T) {
	client.Del("test")
	if n, err := client.HSet("test", "a", "b"); err != nil || n != 1 {
		t.Fatal(n, err)
	}

	if n, err := client.HSet("test", "a", "b"); err != nil || n != 0 {
		t.Fatal(n, err)
	}
}

func TestCache_HSetNX(t *testing.T) {
	client.Del("test")
	if n, err := client.HSetNX("test", "a", "b"); err != nil || n != 1 {
		t.Fatal(n, err)
	}

	if n, err := client.HSetNX("test", "a", "b"); err != nil || n != 0 {
		t.Fatal(n, err)
	}
}

func TestCache_HExists(t *testing.T) {
	client.Del("test")
	if client.Exists("test", "a") != 0 {
		t.Fatal("expect 0")
	}

	if n, err := client.HSet("test", "a", "b"); err != nil || n != 1 {
		t.Fatal(n, err)
	}

	if n, err := client.HExists("test", "a"); err != nil || n != 1 {
		t.Fatal(n, err, "expect 1")
	}
}

func TestCache_HGet(t *testing.T) {
	client.Del("test")
	if val, err := client.HGet("test", "a"); err == nil {
		t.Fatal(val, err)
	}

	if n, err := client.HSet("test", "a", "b"); err != nil || n != 1 {
		t.Fatal(n, err)
	}

	if val, err := client.HGet("test", "a"); err != nil || val != "b" {
		t.Fatal(val, err)
	}

	if val, err := client.HGet("test", "b"); err == nil {
		t.Fatal(val, err)
	}

	if val, err := client.HGet("test1", "b"); err == nil {
		t.Fatal(val, err)
	}
}

func TestCache_HDel(t *testing.T) {
	client.Del("test")
	// 0 nil
	if n, err := client.HDel("test", "a"); err != nil || n != 0 {
		t.Fatal(n, err)
	}

	// 1 nil
	if n, err := client.HSet("test", "a", "b"); err != nil || n != 1 {
		t.Fatal(n, err)
	}

	// 1 nil
	if n, err := client.HDel("test", "a"); err != nil || n != 1 {
		t.Fatal(n, err)
	}

	// 1 nil
	if n, err := client.HDel("test", "a", "b"); err != nil || n != 1 {
		t.Fatal(n, err)
	}


}
