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

// 1.search in local databases
// 2.search in remote redis server
func (c *client) Set(key, value string) (string, error) {
	args := []string{"Set", key, value}

	rc := redis.OKCommand{}
	rc.BaseCommand = redis.NewBaseCommand(args...)

	c.processRedis(&rc)

	return rc.Result()
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

func (c *client) SetNX(key, value string) (string, error) {
	rc := redis.OKCommand{
		BaseCommand: redis.NewBaseCommand("SetNX", key, value),
	}
	c.processRedis(&rc)
	return rc.Result()
}

func (c *client) SetEX(key, value string, expires uint64) (string, error) {
	rc := redis.OKCommand{
		BaseCommand: redis.NewBaseCommand("SetNX", key, value, strconv.FormatUint(expires, 10)),
	}
	c.processRedis(&rc)
	return rc.Result()
}

func (c *client) PSetEX(key, value string, expiresMs uint64) (string, error) {
	rc := redis.OKCommand{
		BaseCommand: redis.NewBaseCommand("SetNX", key, value, strconv.FormatUint(expiresMs, 10)),
	}
	c.processRedis(&rc)
	return rc.Result()
}

func (c *client) GET(key string) (string, error) {
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

func (c *client) GetSet(key, value string) (string, error) {
	rc := redis.StringCommand{
		BaseCommand: redis.NewBaseCommand("GetSet", key, value),
	}
	c.processRedis(&rc)
	return rc.Result()
}

func (c *client) StrLen(key string) (int, error) {
	rc := redis.IntCommand{
		BaseCommand: redis.NewBaseCommand("StrLen", key),
	}
	c.processRedis(&rc)

	return rc.Val, rc.Err
}

func (c *client) Append(key, str string) (int, error) {
	rc := redis.IntCommand{
		BaseCommand: redis.NewBaseCommand("Append", key, str),
	}
	c.processRedis(&rc)

	return rc.Val, rc.Err
}

func (c *client) SetRange(key string, offset int, value string) (int, error) {
	rc := redis.IntCommand{
		BaseCommand: redis.NewBaseCommand("SetRange", key, strconv.Itoa(offset), value),
	}
	c.processRedis(&rc)

	return rc.Val, rc.Err
}

func (c *client) GetRange(key string, start, end int) (string, error) {
	rc := redis.StringCommand{
		BaseCommand: redis.NewBaseCommand("GetRange", key, strconv.Itoa(start), strconv.Itoa(end)),
	}
	c.processRedis(&rc)
	return rc.Result()
}

func (c *client) Incr(key string) (int, error) {
	rc := redis.IntCommand{
		BaseCommand: redis.NewBaseCommand("Incr", key),
	}
	c.processRedis(&rc)

	return rc.Val, rc.Err
}

func (c *client) IncrBy(key string, increment int) (int, error) {
	rc := redis.IntCommand{
		BaseCommand: redis.NewBaseCommand("Incr", key, strconv.Itoa(increment)),
	}
	c.processRedis(&rc)

	return rc.Val, rc.Err
}

func (c *client) IncrByFloat(key string, increment float64) (float64, error) {
	rc := redis.FloatCommand{
		BaseCommand: redis.NewBaseCommand("IncrByFloat", key, strconv.FormatFloat(increment,'f', -1, 64 )),
	}
	c.processRedis(&rc)

	return rc.Val, rc.Err
}

func (c *client) Decr(key string) (int, error) {
	rc := redis.IntCommand{
		BaseCommand: redis.NewBaseCommand("Decr", key),
	}
	c.processRedis(&rc)

	return rc.Val, rc.Err
}

func (c *client) DecrBy(key string, decrement int) (int, error) {
	rc := redis.IntCommand{
		BaseCommand: redis.NewBaseCommand("DecrBy", key, strconv.Itoa(decrement)),
	}
	c.processRedis(&rc)

	return rc.Val, rc.Err
}

func (c *client) MSet(kv map[string]string) (string, error) {
	var args = []string{"MSet"}
	for k, v := range kv{
		args = append(args, k, v)
	}

	rc := redis.OKCommand{
		BaseCommand: redis.NewBaseCommand(args...),
	}

	c.processRedis(&rc)

	return rc.Val, rc.Err
}

func (c *client) MSetNX(kv map[string]string) (int, error) {
	var args = []string{"MSetNX"}
	for k, v := range kv{
		args = append(args, k, v)
	}

	rc := redis.IntCommand{
		BaseCommand: redis.NewBaseCommand(args...),
	}

	c.processRedis(&rc)

	return rc.Val, rc.Err
}

// error(nil) or string
func (c *client) MGet(keys ...string) ([]interface{}, error) {
	rc := redis.InterfaceArrayCommand{
		BaseCommand: redis.NewBaseCommand(append([]string{"MGet"}, keys...)...),
	}
	c.processRedis(&rc)

	return rc.Val, rc.Err
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
