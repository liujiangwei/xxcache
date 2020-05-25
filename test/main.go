package main

import (
	"github.com/liujiangwei/xxcache"
	"log"
)

func main() {
	//client := redis.NewClient(&redis.Options{})
	//client.Get("a")
	//client.Ping().Result()
	//client.config

	cache, err  := xxcache.New(xxcache.Option{
		Addr:"localhost:6379",
	})

	if err != nil{
		log.Println(err)
	}

	cache.Sync()

	cache.Info("a")
	//cache.Sync()
}