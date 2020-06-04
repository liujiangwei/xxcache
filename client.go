package xxcache

import (
	"context"
	"github.com/liujiangwei/xxcache/redis"
	"github.com/sirupsen/logrus"
	"strconv"
)

type client struct {
	cache         Cache
	database      *Database
	databaseIndex int
	connPool      Pool
}

func New(addr string) (*client, error) {
	c := Cache{}
	c.initDatabase(16)

	repl := Replication{
		masterAddr: addr,
	}
	// sync rdb file
	if err := repl.syncWithRedis(); err != nil {
		return nil, err
	}
	// load keys to local cache
	if err := repl.LoadRdbToCache(c); err != nil {
		return nil, err
	}

	// wait for new redis op
	go c.Watch(repl)

	//init client redis conn pool
	pool := Pool{
		addr:     addr,
		capacity: 3,
	}
	if err := pool.Init(); err != nil {
		return nil, err
	}

	cli := client{
		cache:    c,
		connPool: pool,
	}

	if _, err := cli.Select(0); err != nil {
		logrus.Warnln("select db error", err)
		return nil, err
	}

	return &cli, nil
}

func (c *client) processRedis(rc redis.Command) {
	c.connPool.ExecCommand(context.Background(), rc)
}

func (c *client) processCache(cc CacheCommand) {
	entry := c.database.Get(cc.Key())
	cc.Entry(entry)
}

// 1.search in local databases
// 2.search in remote redis server
func (c *client) Set(key, value string) (string, error) {
	args := []string{"Set", key, value}

	rc := redis.OKCommand{}
	rc.BaseCommand = redis.NewBaseCommand(args...)

	c.processRedis(&rc)

	return rc.Result()
}

func (c *client) Get(key string) (string, error) {
	cc := CacheStringCommand{}
	cc.KeyCommand =  NewKeyCommand(key)
	c.processCache(cc)
	if cc.entry != nil{
		logrus.Debugln("from local cache")
		return cc.entry.val, nil
	}

	rc := redis.StringCommand{}
	rc.BaseCommand = redis.NewBaseCommand("Get", key)
	c.processRedis(&rc)

	return rc.Val, rc.Err
}

// 1.search in local databases
// 2.search in remote redis server
func (c *client) Select(index int) (string, error) {
	rc := redis.OKCommand{
		BaseCommand: redis.NewBaseCommand("select", strconv.Itoa(index)),
	}

	c.processRedis(&rc)

	if rc.Err == nil {
		c.database = c.cache.databases[index]
	}

	return rc.Result()
}