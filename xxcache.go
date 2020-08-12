package xxcache

import (
	"github.com/liujiangwei/xxcache/cache"
	"github.com/liujiangwei/xxcache/redis"
	"github.com/sirupsen/logrus"
)

type XXCache interface {
	StringCommand
	//ListCommand
	//xExpiresCommand
}

type Option struct {
	RedisMasterAddr string
	RedisRdbFile    string
	database        int
	LogLevel        logrus.Level
}

type Client struct {
	option Option
	redis  *redis.Redis
	cache  *cache.Database
}

const redisMasterAddr = "127.0.0.1:6379"
const redisRdbFile = "temp.rdb"

func New(option Option) (client Client, err error) {
	client.option = option

	client.cache = new(cache.Database)
	prepareOption(&option)

	repl := redis.Replication{
		MasterAddr: option.RedisMasterAddr,
		RdbFile:    option.RedisRdbFile,
	}

	logrus.SetLevel(option.LogLevel)

	if err = repl.SyncWithRedis(); err != nil {
		return client, err
	}

	go client.handleReplication(repl)

	if err = repl.Load(); err != nil {
		return client, err
	}

	go repl.Ack()

	go repl.WaitForMessage()

	client.redis = redis.NewClient(redis.Option{Addr: option.RedisMasterAddr})

	return client, err
}

func (client *Client) handleReplication(repl redis.Replication) {
	for {
		select {
		case message := <-repl.Messages:
			if message.Database != client.option.database {
				continue
			}

			if err := cache.HandleMessage(client.cache, message.Message); err != nil {
				repl.Err <- err
			}
		case err := <-repl.Err:
			if err != nil {
				logrus.Warnln("repl error", err)
			}
		}
	}
}

func prepareOption(option *Option) {
	if option.RedisMasterAddr == "" {
		option.RedisMasterAddr = redisMasterAddr
	}

	if option.RedisRdbFile == "" {
		option.RedisRdbFile = redisRdbFile
	}
}

func Stop(client *Client) {

}
