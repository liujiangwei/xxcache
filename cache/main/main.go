package main

import (
	"github.com/liujiangwei/xxcache/cache"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)


func main() {
	logrus.Infoln(strconv.Atoi(""))

	return
	test()
	client := cache.New()
	client.HSet("a", "f", "v")
	logrus.Infoln(client.HGet("a", "f"))

	logrus.Print(time.Now().String())
	for i:=0; i < 100000;i++{
		s := strconv.Itoa(i)
		for i:=0; i < 100;i++{
			fv := strconv.Itoa(i)
			client.HSet(s, fv, fv)
		}
	}

	logrus.Print(time.Now().String())
	time.Sleep(time.Second * 10)
}

func test()(n int)  {

	defer logrus.Infoln("n", n)
	n = 2
	return n
}