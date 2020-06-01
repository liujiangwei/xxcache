package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/liujiangwei/xxcache"
	"github.com/sirupsen/logrus"
	"strconv"
)

func main() {
	//testData()

	cache, err := xxcache.New(xxcache.Option{
		Addr: "localhost:6379",
	})

	if err != nil {
		logrus.Warnln(err)
	}

	logrus.Infoln(cache.SyncWithRedis())
	//cache.Sync()
}

func testData() {
	client := redis.NewClient(&redis.Options{})

	for i := 0; i < 10; i++ {
		client.Set(strconv.Itoa(i), i, 0)
	}

	for i := 0; i < 10; i++ {
		client.LPush("list", i)
	}

	for i := 0; i < 10; i++ {
		client.ZAdd("zset", &redis.Z{
			Score:  float64(i),
			Member:  i+1 ,
		})
	}

	for i := 0; i < 10; i++ {
		client.SAdd("set", i)
	}

	for i := 0; i < 10; i++ {
		client.HSet("hash", i+1, i)
	}
}
