package xxcache

import (
	"bufio"
	"context"
	"errors"
	"github.com/cornelk/hashmap"
	"github.com/liujiangwei/xxcache/redis"
	"github.com/liujiangwei/xxcache/redis/rdb"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
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

	//serverInfo map[string]string
	replication Replication

	lock sync.Mutex
}

type Replication struct {
	replicationId           string
	replicationId2          string
	replicationOffset       int
	secondReplicationOffset int
}

const DefaultRdbFile = "tmp.rdb"

type Option struct {
	Addr    string
	RdbFile string
}

func (cache *Cache) syncReplication() (err error) {
	var raw string
	if raw, err = cache.Info("replication"); err != nil {
		return err
	}

	if info, err := redis.ParseInfo(raw); err != nil {
		return err
	} else {
		//cache.replication.replicationId = info.Replication.MasterReplicationId
		cache.replication.replicationId2 = info.Replication.MasterReplicationId2
		cache.replication.replicationOffset = info.Replication.MasterReplOffset
		cache.replication.secondReplicationOffset = info.Replication.SecondReplOffset
	}

	return nil
}

func (cache *Cache) syncDatabase() error {
	conf, err := cache.ConfigGet("databases")
	if err != nil {
		return err
	}

	databases, ok := conf["databases"]
	if !ok {
		return errors.New("failed to get database size from redis")
	}
	num, err := strconv.Atoi(databases)
	if err != nil {
		return errors.New("failed to get database size from redis")
	}
	cache.initializeDatabase(num)

	return nil
}

// Sync this will clear the local cache,so you should call this at first
func (cache *Cache) SyncWithRedis() error {
	logrus.Infoln("start to sync with redis master", cache.option.Addr)
	// sync replication from redis
	if err := cache.syncReplication(); err != nil {
		return err
	}

	// sync database size
	if err := cache.syncDatabase(); err != nil {
		return err
	}

	return cache.connPool.Exec(context.Background(), func(conn *redis.Conn) error {
		pingCmd := NewStringCommand("Ping")
		if message, err := conn.Send(pingCmd.Serialize()); err != nil {
			return err
		} else if message.String() != redis.PONG.String() {
			logrus.Println("ping redis master failed,receive", message.String(), "from", cache.option.Addr)
		}

		// try psync first
		logrus.Infoln("start psync", cache.replication.replicationId, strconv.Itoa(cache.replication.replicationOffset))
		pSyncCmd := NewStringCommand("PSync", cache.replication.replicationId, strconv.Itoa(cache.replication.replicationOffset))
		//FULLRESYNC 022134966dfade095e61201bf10103c5799ac91e 0
		//CONTINUE
		if message, err := conn.Send(pSyncCmd.Serialize()); err != nil {
			return err
		} else if message.String() == "CONTINUE" {
			logrus.Infoln("psync success")
		} else {
			logrus.Infoln("redis master require full resync")
			str := strings.Split(message.String(), " ")
			if len(str) != 3 || str[0] != "FULLRESYNC" {
				logrus.Warnln("sync with redis master failed:", str)
				return errors.New(message.String())
			}
			cache.replication.replicationOffset = 0
			cache.replication.replicationId = str[1]

			var rdb = cache.option.RdbFile
			if rdb == "" {
				logrus.Infoln("use default rdb file", DefaultRdbFile)
				rdb = DefaultRdbFile
			}

			fp, err := os.OpenFile(rdb, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
			if err != nil {
				logrus.Warnln("failed to open rdb file", rdb, err)
				return err
			}

			if _, _, err := conn.ReadWithWriter(fp); err != nil {
				logrus.Warnln("failed to save rdb to file", err)
				return err
			}

			if err := fp.Close(); err != nil {
				logrus.Warnln("failed to close rdb file", err)
			}

			logrus.Infoln("succeed to save the rdb to", rdb)

			if err := cache.loadRdb(rdb); err != nil {
				return err
			}
		}

		return nil
	})
}

func (cache *Cache) loadRdb(filename string) (err error) {
	var file *os.File
	if file, err = os.OpenFile(filename, os.O_RDONLY, os.ModePerm); err != nil {
		return err
	}

	defer func() {
		file.Close()
	}()

	buf := bufio.NewReader(file)
	var version = make([]byte, 9)
	if _, err = buf.Read(version); err != nil {
		return err
	}
	// start to load rdb file
	var opCode uint
	for err == nil {
		// load op code first
		if opCode, err = rdb.LoadOpCode(buf); err != nil {
			return err
		}

		var n int
		switch opCode {
		case rdb.OpCodeExpireTime:
			// 4 byte
			var time = make([]byte, 4)
			if n, err = io.ReadFull(buf, time); err != nil {
				return err
			} else if n != 4 {
				return errors.New("failed to load 4 bytes for OpCodeExpireTime")
			}
			logrus.Infoln("OpCodeExpireTime", string(time))
		case rdb.OpCodeExpireTimeMs:
			var milliTime = make([]byte, 8)
			if n, err = io.ReadFull(buf, milliTime); err != nil {
				return err
			} else if n != 8 {
				return errors.New("failed to load 8 bytes for milliTime")
			}
			logrus.Infoln("OpCodeExpireTime", string(milliTime))
		case rdb.OpCodeFreq:
			var lfu byte
			if lfu, err = buf.ReadByte(); err != nil {
				return err
			}
			logrus.Infoln("OpCodeFreq", lfu)
		case rdb.OpCodeIdle:
			var lru uint64
			if lru, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			logrus.Println("OpCodeIdle", lru)
		case rdb.OpCodeEof:
			logrus.Infoln("success load rdb")
			return nil
		case rdb.OpCodeSelectDB:
			var db uint64
			if db, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			logrus.Println("OpCodeSelectDB", db)
		case rdb.OpCodeResizeDB:
			var dbSize, expiresSize uint64
			if dbSize, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			if expiresSize, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			logrus.Println("OpCodeResizeDB", dbSize, expiresSize)
		case rdb.OpCodeAux:
			var key, value string
			if key, err = rdb.LoadString(buf); err != nil {
				return err
			}
			if value, err = rdb.LoadString(buf); err != nil {
				return err
			}
			switch key {
			case "repl-id":
				cache.replication.replicationId = value
			case "repl-offset":
				if offset, err := strconv.Atoi(value); err != nil {
					return err
				} else {
					cache.replication.replicationOffset = offset
				}
			default:
				logrus.Warnln("unused op code aux", key, value)
			}
		case rdb.OpCodeModuleAux:
			//todo
			logrus.Infoln("OpCodeModuleAux")
		default:
			// this is key value pair
			var key string
			if key, err = rdb.LoadString(buf); err != nil {
				return err
			}
			// opCode is object type
			switch opCode {
			case rdb.TypeString:
				var value string
				if value, err = rdb.LoadString(buf); err != nil {
					return err
				}
				logrus.Infoln("TypeString", key, value)
			case rdb.TypeList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					return err
				}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						return err
					} else {
						logrus.Infoln("TypeList", key, value)
					}
				}
			case rdb.TypeSet:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					return err
				}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						return err
					} else {
						log.Println("TypeSet", key, value)
					}
				}
			case rdb.TypeZSet, rdb.TypeZSet2:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					return err
				}
				var value string
				var score float64
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						return err
					}
					if opCode == rdb.TypeZSet2 {
						if score, err = rdb.LoadBinaryDouble(buf); err != nil {
							return err
						}
						logrus.Infoln("TypeZSet2", key, score, value)
					} else {
						if score, err = rdb.LoadDouble(buf); err != nil {
							return err
						}
						logrus.Infoln("TypeZSet", key, score, value)
					}
				}
			case rdb.TypeHash:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					return err
				}
				var field, value string
				for ; length > 0; length-- {
					if field, err = rdb.LoadString(buf); err != nil {
						return err
					}
					if value, err = rdb.LoadString(buf); err != nil {
						return err
					}

					logrus.Infoln("TypeHash", key, field, value)
				}
			case rdb.TypeListQuickList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					return err
				}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						return err
					}
					logrus.Infoln("TypeListQuickList", key, value)
				}
			case rdb.TypeHashZipMap, rdb.TypeListZipList, rdb.TypeSetIntSet, rdb.TypeZSetZipList, rdb.TypeHashZipList:
				var str string
				if str, err = rdb.LoadString(buf); err != nil {
					return err
				}
				logrus.Infoln("Encode", key, str)
			case rdb.TypeStreamListPacks:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					return err
				}

				for ; length > 0; length-- {
					//if value, err = rdb.LoadString(buf); err != nil {
					//	return err
					//}
					//logrus.Infoln("TypeListQuickList", key, value)
				}
				err = errors.New("TypeStreamListPacks")
			case rdb.TypeModule, rdb.TypeModule2:
				err = errors.New("TypeModule TypeModule2")
			}
		}

	}

	return err
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

	if pool, err := initPool(4, option.Addr); err != nil {
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
	logrus.Debugln("PSync", id, offset)

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
