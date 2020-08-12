package cache

import (
	"errors"
	"github.com/liujiangwei/xxcache/redis"
	"github.com/sirupsen/logrus"
	"strings"
)

func HandleMessage(cache *Cache, message redis.Message) (err error) {
	logrus.Infoln(message)
	messages, ok := message.(redis.ArrayMessage)
	if !ok {
		return errors.New("error redis message, need array message")
	}

	if len(messages) == 0 {
		return errors.New("empty message")
	}

	switch strings.ToLower(messages[0].String()) {
	case "set":
		_, err = cache.Set(messages[1].String(), messages[2].String())
	case "rpush":
		if len(messages) < 2 {
			return errors.New("rpush command args error")
		}

		var values []string
		for i := 2; i < len(messages); i++ {
			values = append(values, messages[i].String())
		}
		_, err = cache.RPush(messages[1].String(), values...)
	case "hmset":
		if len(messages) < 2 {
			return errors.New("hmset command args error")
		}

	default:
		return errors.New("unknown handled command," + message.Serialize())
	}

	return err
}
