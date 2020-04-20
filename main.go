package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/liujiangwei/xxcache/cache"
	"log"
	"sync"
)

func main() {

	server := cache.Cache{}

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println(server.Sync(":6379"))

	return
	//log.Println(server.Get("a"))
	num := 100
	wg := sync.WaitGroup{}
	wg.Add(num)

	for i :=0; i < num; i++{
		go func() {
			defer wg.Done()
			log.Println(server.Incr("int"))
			log.Println(server.Incr("int"))
		}()
	}

	wg.Wait()
	log.Println(server.Incr("int"))

	client := redis.NewClient(&redis.Options{})
	client.Set("a", "b", 0).Result()
	client.Set("a", server, 0)
	client.Get("a").Result()
}
