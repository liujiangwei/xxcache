package xxcache

import (
	"github.com/cornelk/hashmap"
	"github.com/sirupsen/logrus"
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
	connPool Pool
	databases []*Database
	sync.RWMutex
}

type MessageHandleFunc func(cache *Cache)
func (cache *Cache) initDatabase(size int) {
	cache.databases = make([]*Database, size)

	for i := 0; i < size; i++ {
		cache.databases[i] = &Database{dict: hashmap.HashMap{}}
	}
}

func (cache *Cache) initPool (capacity int, addr string) error {
	pool := Pool{
		addr:addr,
		capacity:capacity,
	}

	if err := pool.Init(); err != nil{
		return err
	}
	cache.connPool = pool

	return nil
}

func (cache *Cache) SelectDatabase(index int) *Database{
	if index < 0 || index >= len(cache.databases){
		return nil
	}

	return cache.databases[index]
}