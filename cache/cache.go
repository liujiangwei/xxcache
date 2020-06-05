package cache

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cornelk/hashmap"
	"github.com/liujiangwei/xxcache/database"
	"github.com/liujiangwei/xxcache/redis"
	"github.com/liujiangwei/xxcache/redis/intset"
	"github.com/liujiangwei/xxcache/redis/rdb"
	"github.com/liujiangwei/xxcache/redis/ziplist"
	"github.com/liujiangwei/xxcache/redis/zipmap"
	skiplist "github.com/sean-public/fast-skiplist"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

type Cache struct {
	databases []*database.Database
	sync.RWMutex
}

func (cache *Cache) InitDatabase(size int) {
	cache.databases = make([]*database.Database, size)

	for i := 0; i < size; i++ {
		cache.databases[i] = &database.Database{Dict: hashmap.HashMap{}}
	}
}

func (cache *Cache) SelectDatabase(index int) *database.Database{
	if index < 0 || index >= len(cache.databases){
		return nil
	}

	return cache.databases[index]
}

// Sync this will clear the local cache,so you should call this at first
func (cache *Cache)Watch(repl redis.Replication) {
	var ack = time.NewTicker(time.Second)
	for {
		select {
		case err := <-repl.Err:
			logrus.Warnln("replication", err)
		case message := <-repl.Messages:
			cache.handel(message)
		case <-ack.C:
			if err := repl.Ack(); err != nil {
				logrus.Warnln("replication Ack", err)
			}
		}
	}
}

func (cache *Cache)handel(message redis.Message) {
	logrus.Infoln(message.String())
	messages,ok := message.(redis.ArrayMessage)
	if !ok || len(messages) == 0{
		logrus.Warnln("error")
		return
	}

	var args []string
	for _,m := range messages{
		args = append(args, m.String())
	}
	if len(args) == 0{
		return
	}

	switch args[0] {

	}
	//cmd := args[0]
}


func (cache *Cache) LoadReplication(repl *redis.Replication) (err error) {
	db := cache.SelectDatabase(0)

	var file *os.File
	if file, err = os.OpenFile(redis.RdbFile, os.O_RDONLY, os.ModePerm); err != nil {
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
			var expireS = make([]byte, 4)
			if n, err = io.ReadFull(buf, expireS); err != nil {
				return err
			} else if n != 4 {
				return errors.New("failed to load 4 bytes for OpCodeExpireTime")
			}
			expiresTime = uint64(binary.LittleEndian.Uint32(expireS) * 1000)
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
				db = cache.SelectDatabase(int(index))
				if db == nil{
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
					entry := database.StringEntry{Val: value}
					db.Set(key, entry)
					logrus.Debugln("TypeString", key, value)
				}
			case rdb.TypeList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeList length", err)
					return err
				}

				entry := database.ListEntry{}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeList value", err)
						return err
					} else {
						entry.Val = append(entry.Val, value)
					}
				}
				db.Set(key, entry)
				logrus.Debugln("TypeList", key, value)
			case rdb.TypeSet:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeSet length", err)
					return err
				}

				entry := database.SetEntry{
					Val:hashmap.New(uintptr(int(length))),
				}
				for ; length > 0; length-- {
					if value, err := rdb.LoadString(buf); err != nil {
						return err
					} else {
						entry.Val.Set(value, true)
					}
				}
				db.Set(key, entry)
				logrus.Debugln("TypeSet", key, entry.Val.String())
			case rdb.TypeZSet, rdb.TypeZSet2:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeZSet, TypeZSet2 length", err)
					return err
				}

				entry := database.ZSetEntry{Val: skiplist.New()}
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

					entry.Val.Set(score, value)
				}
				db.Set(key, entry)
			case rdb.TypeHash:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeHash length", err)
					return err
				}
				entry := database.HashEntry{Val: hashmap.New(uintptr(int(length)))}
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
					entry.Val.Set(field, value)
				}
				db.Set(key, value)
			case rdb.TypeListQuickList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeListQuickList length", err)
					return err
				}

				entry := database.ListEntry{}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeListQuickList value", err)
						return err
					}
					entry.Val = append(entry.Val, value)
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
					entry := database.HashEntry{
						Val: hashmap.New(uintptr(10)),
					}
					for field, value := range hash{
						entry.Val.Set(field, value)
					}
					db.Set(key, entry)
					logrus.Infoln("TypeHashZipMap", key, hash, entry.Val.String())
				case rdb.TypeSetIntSet:
					entry := database.SetEntry{Val: hashmap.New(10)}
					set := intset.Load(str)
					for _, v := range set{
						entry.Val.Set(v, true)
					}
					logrus.Infoln("TypeSetIntSet", key, set)
				case rdb.TypeZSetZipList:
					// hash member => score
					entry := database.ZSetEntry{Val: skiplist.New()}
					list := ziplist.Load(str)
					size := len(list) / 2
					for i := 0; i < size; i++ {
						if score, err := strconv.ParseFloat(list[2*i+1], 64); err != nil{
							err = errors.New(fmt.Sprintf("zset format error,%s", err))
							logrus.Warnln("TypeZSetZipList parse Err", err)
						}else{
							entry.Val.Set(score, list[2*i])
						}
					}

					db.Set(key, entry)
				case rdb.TypeHashZipList:
					entry := database.HashEntry{
						Val: hashmap.New(uintptr(10)),
					}
					list := ziplist.Load(str)
					size := len(list) / 2
					for i := 0; i < size; i++ {
						entry.Val.Set(list[2*i],  list[2*i+1])
						logrus.Infoln("TypeHashZipList", "field=>value", key, list[2*i], list[2*i+1])
					}
					db.Set(key, entry)
				case rdb.TypeListZipList:
					entry := database.ListEntry{}
					list := ziplist.Load(str)
					for i := 0; i < len(list); i++ {
						entry.Val = append(entry.Val, list[i])
						logrus.Infoln("TypeListZipList", key, list[i])
					}
					db.Set(key, entry)
				}
			case rdb.TypeStreamListPacks:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					return err
				}

				for ; length > 0; length-- {
					//if value, Err = rdb.LoadString(buf); Err != nil {
					//	return Err
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
