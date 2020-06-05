package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/liujiangwei/xxcache"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func main() {

	//testData()
	client , err := xxcache.New("localhost:6379")
	if err != nil {
		logrus.Fatalln(err)
	}

	logrus.Infoln(client.Get("aae"))

	time.Sleep(time.Second * 60)
	//cache.Sync()
}

func testData() {
	client := redis.NewClient(&redis.Options{})
	logrus.Fatalln(client.Get("a").Result())
	for i := 0; i < 1; i++ {
		// string
		client.Set(strconv.Itoa(i), i, 0).Result()
		client.Set("string" + strconv.Itoa(i), i, 0)

		//list
		//client.Del("list")
		client.LPush("list", i)
		//client.Del("list-string")
		client.LPush("list-string", "string-" + strconv.Itoa(i)+"-a")

		// [29 0 0 0 19 0 0 0 2 0 0 7 102 105 101 108 100 45 48 9 7 118 97 108 117 101 45 48 255]
		//set
		//client.Del("set")
		client.SAdd("set", i)
		client.SAdd("set-100", i* 100)
		//client.Del("set-100")
		client.SAdd("set-10000", i * 10000)
		//client.Del("set-1000000")
		client.SAdd("set-1000000", i * 1000000)

		//sorted set
		//client.Del("zset")
		client.ZAdd("zset", &redis.Z{
			Score:  float64(i),
			Member:  i+1 ,
		})

		// hash
		client.HSet("hash", i, i+1)
		client.HSet("hash-string", "field-" + strconv.Itoa(i), "value-" + strconv.Itoa(i))
	}
}
