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
	RedisRdbFile string
	database int
}

type Client struct {
	option Option
	redis *redis.Redis
	cache *cache.Cache
}

const redisMasterAddr = "127.0.0.1:6379"
const redisRdbFile = "temp.rdb"
func New(option Option) (client Client, err error){
	client.option = option

	client.cache = new(cache.Cache)
	prepareOption(&option)

	repl := redis.Replication{
		MasterAddr:option.RedisMasterAddr,
		RdbFile:option.RedisRdbFile,
	}

	if err = repl.SyncWithRedis(); err != nil{
		return client, err
	}

	go client.handleMessage(repl)

	go client.handleError(repl)

	if err = repl.Load(); err != nil{
		return client, err
	}

	go repl.WaitForMessage()

	client.redis = redis.New(redis.Options{Addr:option.RedisMasterAddr})

	return client, err
}


func (client *Client) handleMessage(repl redis.Replication) {
	for {
		select {
		case message := <-repl.Messages:
			if message.Database != client.option.database{
				continue
			}

			if err :=  cache.HandleMessage(client.cache, message.Message); err != nil{
				repl.Err <- err
			}
		}
	}
}

func (client *Client)handleError(repl redis.Replication)  {
	for {
		err := <- repl.Err
		if err != nil{
			logrus.Warnln("repl error", err)
		}
	}
}


func prepareOption(option *Option) {
	if option.RedisMasterAddr == ""{
		option.RedisMasterAddr = redisMasterAddr
	}

	if option.RedisRdbFile == ""{
		option.RedisRdbFile = redisRdbFile
	}
}

func Stop(client *Client) {

}