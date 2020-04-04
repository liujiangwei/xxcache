package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/liujiangwei/xxcache/service"
	"log"
)

func main() {
	client := redis.NewClient(&redis.Options{})
	client.Set("a", "b", 0).Result()

	server := service.Server{}
	log.Println(server.Start(":6379"))
}
