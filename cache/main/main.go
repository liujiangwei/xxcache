package main

import (
	"github.com/liujiangwei/xxcache/cache"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func test()  {
	
}

func main() {
	client := cache.New()
	client.Set("a", "aaa")
	logrus.Infoln(client.Get("a"))

	return
	logrus.Print(time.Now().String())
	for i:=0; i < 10000000;i++{
		s := strconv.Itoa(i)
		client.SetEX(s, s, 1)
	}

	logrus.Print(time.Now().String())
	time.Sleep(time.Second * 10)
}
