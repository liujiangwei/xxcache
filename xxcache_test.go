package xxcache

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestNew(t *testing.T) {
	option := Option{
		RedisMasterAddr: "127.0.0.1:6379",
		RedisRdbFile:    "/tmp/redis.rdb",
		database:        0,
		LogLevel:        logrus.WarnLevel,
	}

	_, err := New(option)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Get(t *testing.T) {
	option := Option{
		RedisMasterAddr: "127.0.0.1:6379",
		RedisRdbFile:    "/tmp/redis.rdb",
		database:        0,
		LogLevel:        logrus.WarnLevel,
	}

	client, err := New(option)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := client.Set("a", "this is value for a"); err != nil {
		t.Fatal("failed to set a", err)
	}

	if val, err := client.Get("a"); err != nil || val != "this is value for a" {
		t.Fatal("failed to get a", val, err)
	}
}
