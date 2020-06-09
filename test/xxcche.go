package main

import (
	"github.com/liujiangwei/xxcache"
	"github.com/sirupsen/logrus"
	"time"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	cache, err := xxcache.New(xxcache.Option{
		RedisMasterAddr:"127.0.0.1:6379",
		RedisRdbFile:"tmp.rdb",
	})

	if err != nil{
		logrus.Fatal(err)
	}

	cache.Get("string-1")
	time.Sleep(time.Minute  * 2)
}
