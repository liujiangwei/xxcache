package cache

import (
	"errors"
	"github.com/liujiangwei/xxcache/redis"
	"github.com/sirupsen/logrus"
	"strings"
)

func HandleMessage(cache *Cache, message redis.Message) error{
	logrus.Infoln(message)
	messages, ok := message.(redis.ArrayMessage)
	if !ok {
		return errors.New("error redis message, need array message")
	}

	if len(messages) == 0{
		return errors.New("empty message")
	}

	switch strings.ToLower(messages[0].String()) {
	case "set":
		if _, err := cache.Set(messages[1].String(), messages[2].String()); err != nil{
			return err
		}

		logrus.Debugln(cache.Get(messages[1].String()))
	default:
		return errors.New("unknown handled command,"+ message.String())
	}

	return nil
}