package xxcache

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cornelk/hashmap"
	"github.com/liujiangwei/xxcache/redis"
	"github.com/liujiangwei/xxcache/redis/intset"
	"github.com/liujiangwei/xxcache/redis/rdb"
	"github.com/liujiangwei/xxcache/redis/ziplist"
	"github.com/liujiangwei/xxcache/redis/zipmap"
	"github.com/sean-public/fast-skiplist"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

type Replication struct {
	masterAddr string

	Id     string
	Offset int

	messages chan redis.Message
	err      chan error
	stop     chan struct{}

	database *Database

	conn *redis.Conn

	sync.Mutex
}

const DefaultReplicationMessages = 10000

func (repl *Replication) Stop() {
}

func (repl *Replication) syncWithRedis() error {
	repl.messages = make(chan redis.Message, DefaultReplicationMessages)
	repl.err = make(chan error, 1)

	if repl.masterAddr == "" {
		return errors.New("redis master addr is required")
	}

	if conn, err := redis.Connect(repl.masterAddr); err != nil {
		err = errors.New("failed to connect to master," + err.Error())
		return err
	} else {
		repl.conn = conn
	}

	if err := repl.Ping(); err != nil {
		logrus.Warn("replication error", err)
		return err
	}

	if err := repl.Sync(); err != nil {
		logrus.Warnln("replication error", err)
		return err
	}

	// wait for message from master
	go repl.WaitForMessage()

	return nil
}

func (repl *Replication) WaitForMessage() {
	for {
		if message, err := repl.conn.Recv(); err != nil {
			close(repl.messages)
			repl.err <- err
			return
		} else {
			repl.messages <- message
		}
	}
}

func (repl *Replication) Ack() error {
	cmd := redis.StringCommand{
		BaseCommand: redis.NewBaseCommand("ReplConf", "Ack", strconv.Itoa(repl.Offset)),
	}
	if err := repl.conn.Send(cmd.Serialize()); err != nil {
		err = errors.New("failed send Ack to master," + err.Error())
		return err
	}

	return nil
}

func (repl *Replication) Ping() error {
	cmd := redis.StringCommand{
		BaseCommand: redis.NewBaseCommand("Ping"),
	}
	if message, err := repl.conn.SendAndWaitReply(cmd.Serialize()); err != nil {
		return err
	} else if message.String() != redis.PONG.String() {
		err = errors.New(fmt.Sprintf("Ping, receive %s from %s", message.String(), repl.masterAddr))
		return err
	}

	return nil
}

const rdbFile = "./tmp.rdb"

func (repl *Replication) Sync() (err error) {
	cmd := redis.StringCommand{
		BaseCommand: redis.NewBaseCommand("SYNC"),
	}

	if err = repl.conn.Send(cmd.Serialize()); err != nil {
		err = errors.New("Sync error," + err.Error())
		return err
	}

	var fp *os.File
	if fp, err = os.OpenFile(rdbFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm); err != nil {
		err = errors.New("failed to open rdb file," + err.Error())
		return err
	}

	if protocol, length, err := repl.conn.ReadWithWriter(fp); err != nil {
		err = errors.New("failed to save rdb file," + err.Error())
		return err
	} else {
		logrus.Infoln("redis rdb file", redis.Protocol(protocol), length)
	}

	return nil
}

func (repl *Replication) Stat(duration time.Duration) {
	for range time.NewTicker(duration).C {
		logrus.Infoln(fmt.Sprintf("Stat Id [%s] Offset[%d]", repl.Id, repl.Offset))
	}
}

func (repl *Replication) SetReplicationId(id string) {
	repl.Id = id
}

func (repl *Replication) SetReplicationOffset(offset int) {
	repl.Offset = offset
}

func (repl *Replication) LoadRdbToCache(cache Cache) (err error) {
	repl.database = cache.SelectDatabase(0)

	var file *os.File
	if file, err = os.OpenFile(rdbFile, os.O_RDONLY, os.ModePerm); err != nil {
		return errors.New("failed to open rdb file," + err.Error())
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
	var expiresTime uint64
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
			expiresTime = uint64(binary.LittleEndian.Uint32(time) * 1000)
			logrus.Infoln("OpCodeExpireTime", expiresTime)

		case rdb.OpCodeExpireTimeMs:
			var milliTime = make([]byte, 8)
			if n, err = io.ReadFull(buf, milliTime); err != nil {
				return err
			} else if n != 8 {
				return errors.New("failed to load 8 bytes for milliTime")
			}
			expiresTime = binary.LittleEndian.Uint64(milliTime)
			logrus.Infoln("OpCodeExpireTimeMs", expiresTime)
		case rdb.OpCodeFreq:
			var lfu byte
			if lfu, err = buf.ReadByte(); err != nil {
				return err
			}

			logrus.Infoln("OpCodeFreq", lfu)
			continue
		case rdb.OpCodeIdle:
			var lru uint64
			if lru, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			logrus.Println("OpCodeIdle", lru)
			continue
		case rdb.OpCodeEof:
			logrus.Infoln("rdb done!!!")

			return nil
		case rdb.OpCodeSelectDB:
			if index, _, err := rdb.LoadLen(buf); err != nil {
				return errors.New("OpCodeSelectDB error, " + err.Error() )
			}else{
				repl.database = cache.SelectDatabase(int(index))
				if repl.database == nil{
					err = errors.New(fmt.Sprintf("cache db size error[%d]", index))
					logrus.Warnln(err)
					return err
				}

				logrus.Println("OpCodeSelectDB", index)
			}

			continue
		case rdb.OpCodeResizeDB:
			var dbSize, expiresSize uint64
			if dbSize, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			if expiresSize, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			logrus.Println("OpCodeResizeDB ignored", dbSize, expiresSize)
			continue
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
				repl.SetReplicationId(value)
			case "repl-offset":
				if offset, err := strconv.Atoi(value); err != nil {
					return err
				} else {
					repl.SetReplicationOffset(offset)
				}
			default:
				logrus.Warnln("unused op code aux", key, value)
			}
			continue
		case rdb.OpCodeModuleAux:
			var moduleId, whenOpCode, when, eof uint64
			if moduleId, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			if whenOpCode, _, err = rdb.LoadLen(buf); err != nil {
				return err
			} else if int(whenOpCode) != rdb.ModuleOpCodeUint {
				return errors.New("ModuleOpCodeUint error")
			}

			if when, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}

			if eof, _, err = rdb.LoadLen(buf); err != nil {
				return err
			} else if int(eof) != rdb.ModuleOpCodeEof {
				return errors.New("ModuleOpCodeEof error")
			}

			logrus.Warnln("OpCodeModuleAux ignored", moduleId, whenOpCode, when, eof)
		default:
			// this is key value pair
			var key string
			if key, err = rdb.LoadString(buf); err != nil {
				logrus.Warnln("failed to load key", err)
				return err
			} else if key == "" {
				logrus.Warnln("failed to load key, key is empty", err)
				return errors.New("empty key")
			}

			// opCode is object type
			switch opCode {
			case rdb.TypeString:
				if value, err := rdb.LoadString(buf); err != nil {
					err = errors.New("failed to load TypeString," + err.Error())
					logrus.Warnln("rdb TypeString", err)
					return err
				}else{
					entry := StringEntry{val:value}
					repl.database.Set(key, entry)
					logrus.Debugln("TypeString", key, value)
				}
			case rdb.TypeList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeList length", err)
					return err
				}

				entry := ListEntry{}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeList value", err)
						return err
					} else {
						entry.val = append(entry.val, value)
					}
				}
				repl.database.Set(key, entry)
				logrus.Debugln("TypeList", key, value)
			case rdb.TypeSet:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeSet length", err)
					return err
				}

				entry := SetEntry{
					val:hashmap.New(uintptr(int(length))),
				}
				for ; length > 0; length-- {
					if value, err := rdb.LoadString(buf); err != nil {
						return err
					} else {
						entry.val.Set(value, true)
					}
				}
				repl.database.Set(key, entry)
				logrus.Debugln("TypeSet", key, entry.val.String())
			case rdb.TypeZSet, rdb.TypeZSet2:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeZSet, TypeZSet2 length", err)
					return err
				}

				entry := ZSetEntry{val:skiplist.New()}
				var value string
				var score float64
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeZSet, TypeZSet2 value", err)
						return err
					}
					if opCode == rdb.TypeZSet2 {
						if score, err = rdb.LoadBinaryDouble(buf); err != nil {
							logrus.Warnln("failed to load TypeZSet, TypeZSet2 value TypeZSet2", err)
							return err
						}
						logrus.Infoln("TypeZSet2", key, score, value)
					} else {
						if score, err = rdb.LoadDouble(buf); err != nil {
							logrus.Warnln("failed to load TypeZSet, TypeZSet2 value TypeZSet2", err)

							return err
						}
						logrus.Infoln("TypeZSet", key, score, value)
					}

					entry.val.Set(score, value)
				}
				repl.database.Set(key, entry)
			case rdb.TypeHash:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeHash length", err)
					return err
				}
				entry := HashEntry{val:hashmap.New(uintptr(int(length)))}
				var field, value string
				for ; length > 0; length-- {
					if field, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeHash field", err)
						return err
					}
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeHash value", err)
						return err
					}

					logrus.Infoln("TypeHash", key, field, value)
					entry.val.Set(field, value)
				}
				repl.database.Set(key, value)
			case rdb.TypeListQuickList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeListQuickList length", err)
					return err
				}

				entry := ListEntry{}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeListQuickList value", err)
						return err
					}
					entry.val = append(entry.val, value)
					logrus.Infoln("TypeListQuickList", key, ziplist.Load(value))
				}

			case rdb.TypeHashZipMap, rdb.TypeListZipList, rdb.TypeSetIntSet, rdb.TypeZSetZipList, rdb.TypeHashZipList:
				var str string
				if str, err = rdb.LoadString(buf); err != nil {
					logrus.Warnln("failed to load TypeHashZipMap... length", key, err)
					return err
				}

				switch opCode {
				case rdb.TypeHashZipMap:
					hash := zipmap.Load(str)
					entry := HashEntry{
						val: hashmap.New(uintptr(10)),
					}
					for field, value := range hash{
						entry.val.Set(field, value)
					}
					repl.database.Set(key, entry)
					logrus.Infoln("TypeHashZipMap", key, hash, entry.val.String())
				case rdb.TypeSetIntSet:
					entry := SetEntry{val:hashmap.New(10)}
					set := intset.Load(str)
					for _, v := range set{
						entry.val.Set(v, true)
					}
					logrus.Infoln("TypeSetIntSet", key, set)
				case rdb.TypeZSetZipList:
					// hash member => score
					entry := ZSetEntry{val:skiplist.New()}
					list := ziplist.Load(str)
					size := len(list) / 2
					for i := 0; i < size; i++ {
						if score, err := strconv.ParseFloat(list[2*i+1], 64); err != nil{
							err = errors.New(fmt.Sprintf("zset format error,%s", err))
							logrus.Warnln("TypeZSetZipList parse err", err)
						}else{
							entry.val.Set(score, list[2*i])
						}
					}

					repl.database.Set(key, entry)
				case rdb.TypeHashZipList:
					entry := HashEntry{
						val: hashmap.New(uintptr(10)),
					}
					list := ziplist.Load(str)
					size := len(list) / 2
					for i := 0; i < size; i++ {
						entry.val.Set(list[2*i],  list[2*i+1])
						logrus.Infoln("TypeHashZipList", "field=>value", key, list[2*i], list[2*i+1])
					}
					repl.database.Set(key, entry)
				case rdb.TypeListZipList:
					entry := ListEntry{}
					list := ziplist.Load(str)
					for i := 0; i < len(list); i++ {
						entry.val = append(entry.val, list[i])
						logrus.Infoln("TypeListZipList", key, list[i])
					}
					repl.database.Set(key, entry)
				}
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
				err = errors.New("TypeStreamListPacks unsupported now")
			case rdb.TypeModule, rdb.TypeModule2:
				err = errors.New("TypeModule TypeModule2 unsupported now")
			}

			if expiresTime > 0 {
				logrus.Infoln("key expires", key, expiresTime)
				expiresTime = 0
			}
		}
	}

	return err
}
