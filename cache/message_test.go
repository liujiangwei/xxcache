package cache

import (
	"github.com/liujiangwei/xxcache/redis"
	"testing"
)

func TestHandleMessage(t *testing.T) {
	cache := new(Database)
	for _, cmd := range testCommandList() {
		if err := HandleMessage(cache, cmd.Serialize()); err != nil {
			t.Fatal(cmd.Serialize(), err)
		}
	}
}

func testCommandList() []redis.BaseCommand {
	return []redis.BaseCommand{
		redis.NewBaseCommand("set", "a", "aa"),
		redis.NewBaseCommand("set", "a", "aa"),
		redis.NewBaseCommand("set", "a", "aa"),
	}
}
