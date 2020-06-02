package xxcache

import (
	"context"
	"github.com/cornelk/hashmap"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
	customFormatter := new(logrus.TextFormatter)
	customFormatter.FullTimestamp = true                    // 显示完整时间
	customFormatter.TimestampFormat = "2006-01-02 15:04:05" // 时间格式
	customFormatter.DisableTimestamp = false                // 禁止显示时间
	customFormatter.DisableColors = false                   // 禁止颜色显示
	logrus.SetFormatter(customFormatter)
}

type Cache struct {
	option   Option
	connPool *Pool

	database []*Database
	cdb      *Database

	lock sync.Mutex
}

const DefaultRdbFile = "tmp.rdb"

type Option struct {
	Addr    string
	RdbFile string
}


func (cache *Cache) initializeDatabase(num int) {
	cache.database = make([]*Database, num)

	for i := 0; i < num; i++ {
		cache.database[i] = &Database{dict: hashmap.HashMap{}}
	}
}

func New(option Option) (*Cache, error) {
	cache := Cache{
		option: option,
	}

	if pool, err := initPool(1, option.Addr); err != nil {
		return nil, err
	} else {
		cache.connPool = pool
	}

	return &cache, nil
}

func (cache Cache) Process(command Command) {
	cache.connPool.ExecCommand(context.Background(), command)
}

func (cache Cache) Ping() (string, error) {
	cmd := NewStringCommand("Ping")

	cache.Process(&cmd)

	return cmd.val, cmd.err
}

func (cache Cache) ReplConf(option string, val string) (string, error) {
	cmd := NewStringCommand("ReplConf", option, val)

	cache.Process(&cmd)

	return cmd.val, cmd.err
}

func (cache Cache) PSync(id string, offset int) (string, error) {
	cmd := NewStringCommand("PSync", id, strconv.Itoa(offset))

	cache.Process(&cmd)

	return cmd.val, cmd.err
}

func (cache Cache) Info(sections ...string) (string, error) {
	args := append([]string{"INFO"}, sections...)

	cmd := NewStringCommand(args...)
	cache.Process(&cmd)

	return cmd.val, cmd.err
}

func (cache Cache) ConfigGet(sections ...string) (map[string]string, error) {
	args := append([]string{"CONFIG", "GET"}, sections...)

	cmd := NewStringStringCommand(args...)
	cache.Process(&cmd)

	return cmd.val, cmd.err
}

func (cache Cache) Set(key, value string) (string, error){
	cmd := NewStringCommand("SET", key, value)
	cache.Process(&cmd)

	return cmd.val, cmd.err
}