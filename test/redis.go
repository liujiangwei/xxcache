package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

func main() {
	client := redis.NewClient(&redis.Options{})
	logrus.Infoln(client.Get("aaa").Result())
}
