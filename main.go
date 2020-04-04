package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/liujiangwei/xxcache/service"
	"log"
)

func main() {
	server := service.Server{}

	client := redis.NewClient(&redis.Options{})
	client.Set("a", "b", 0).Result()
	client.Set("a", server, 0)
	client.Get("a").Result()

	if err := server.Start(); err != nil{
		log.Fatal(err)
	}

	if v, err := server.Get("a"); err != nil{
		log.Println("ok, a is not set, return err")
	}else{
		log.Fatal("a", "=", v)
	}

	server.Set("a", "b")

	if v, err := server.Get("a"); err == nil{
		log.Println("a", "=", v)
	}else{
		log.Fatal("wrong, a is set, should return a")
	}
	log.Fatal(server.Listen(":6380"))
}
