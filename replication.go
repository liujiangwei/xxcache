package xxcache

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/liujiangwei/xxcache/redis"
	"github.com/liujiangwei/xxcache/redis/intset"
	"github.com/liujiangwei/xxcache/redis/rdb"
	"github.com/liujiangwei/xxcache/redis/ziplist"
	"github.com/liujiangwei/xxcache/redis/zipmap"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

type Replication struct {
	masterAddr              string
	replicationId           string
	replicationId2          string
	replicationOffset       int
	secondReplicationOffset int
	stat                    bool
	messages                chan redis.Message
	err                     chan error

	conn *redis.Conn

	sync.Mutex
}

const DefaultReplicationMessages = 10000

func (repl *Replication) Stop() error {
	return nil
}

func (repl *Replication) Start() error {
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

	if err := repl.ping(); err != nil {
		logrus.Warn("replication error", err)
		return err
	}

	if err := repl.sync(); err != nil {
		logrus.Warnln("replication error", err)
	}

	// wait for message from master
	go func() {
		for {
			if message, err := repl.conn.Recv(); err != nil {
				close(repl.messages)
				repl.err <- err
				return
			} else {
				repl.messages <- message
			}
		}
	}()

	if repl.stat {
		repl.Stat(time.Second * 5)
	}

	return nil
}

func (repl *Replication) ack() error {
	cmd := NewStringCommand("ReplConf", "ack", strconv.Itoa(repl.replicationOffset))
	if err := repl.conn.Send(cmd.Serialize()); err != nil {
		err = errors.New("failed send ack to master," + err.Error())
		return err
	}

	return nil
}

func (repl *Replication) ping() error {
	pingCmd := NewStringCommand("Ping")
	if message, err := repl.conn.SendAndWaitReply(pingCmd.Serialize()); err != nil {
		return err
	} else if message.String() != redis.PONG.String() {
		err = errors.New(fmt.Sprintf("ping, receive %s from %s", message.String(), repl.masterAddr))
		return err
	}

	return nil
}

const rdbFile = "./tmp.rdb"

func (repl *Replication) sync() (err error) {
	cmd := NewStringCommand("Sync")
	if err = repl.conn.Send(cmd.Serialize()); err != nil {
		err = errors.New("sync error," + err.Error())
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
		return repl.loadRdb(rdbFile)
	}
}

func (repl *Replication) Stat(duration time.Duration) {
	for range time.NewTicker(duration).C {
		logrus.Infoln(fmt.Sprintf("Stat replicationId [%s] replicationOffset[%d]", repl.replicationId, repl.replicationOffset))
	}
}

func (repl *Replication) SetReplicationId(id string) {
	repl.replicationId = id
}

func (repl *Replication) SetReplicationOffset(offset int) {
	repl.replicationOffset = offset
}

func (repl *Replication) loadRdb(filename string) (err error) {
	var file *os.File
	if file, err = os.OpenFile(filename, os.O_RDONLY, os.ModePerm); err != nil {
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
			logrus.Infoln("success load rdb")
			return nil
		case rdb.OpCodeSelectDB:
			var db uint64
			if db, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			logrus.Println("OpCodeSelectDB", db)
			continue
		case rdb.OpCodeResizeDB:
			var dbSize, expiresSize uint64
			if dbSize, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			if expiresSize, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			logrus.Println("OpCodeResizeDB", dbSize, expiresSize)
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

			logrus.Warnln("OpCodeModuleAux", moduleId, whenOpCode, when, eof)
			return errors.New("OpCodeModuleAux is unsupported")
		default:
			// this is key value pair
			var key string
			if key, err = rdb.LoadString(buf); err != nil {
				logrus.Warnln("failed to load key", err)
				return err
			}else if key == ""{
				logrus.Warnln("failed to load key, key is empty", err)
				return errors.New("empty key")
			}

			// opCode is object type
			switch opCode {
			case rdb.TypeString:
				var value string
				if value, err = rdb.LoadString(buf); err != nil {
					logrus.Warnln("failed to load TypeString", err)
					return err
				}
				logrus.Infoln("TypeString", key, value)
			case rdb.TypeList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeList length", err)

					return err
				}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeList value", err)
						return err
					} else {
						logrus.Infoln("TypeList", key, value)
					}
				}
			case rdb.TypeSet:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeSet length", err)
					return err
				}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						return err
					} else {
						logrus.Infoln("TypeSet", key, value)
					}
				}
			case rdb.TypeZSet, rdb.TypeZSet2:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeZSet, TypeZSet2 length", err)
					return err
				}
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
				}
			case rdb.TypeHash:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeHash length", err)
					return err
				}
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
				}
			case rdb.TypeListQuickList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeListQuickList length", err)
					return err
				}
				var value string
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeListQuickList value", err)
						return err
					}
					logrus.Infoln("TypeListQuickList", length, key, ziplist.Load(value))
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
					logrus.Infoln("TypeHashZipMap", key, hash)
				case rdb.TypeSetIntSet:
					set := intset.Load(str)
					logrus.Infoln("TypeSetIntSet", key, set)
				case rdb.TypeZSetZipList:
					// hash member => score
					list := ziplist.Load(str)
					size := len(list) / 2
					for i := 0; i < size; i++ {
						logrus.Infoln("TypeZSetZipList", "member=>score", key, list[2*i], list[2*i+1])
					}
				case rdb.TypeHashZipList:
					list := ziplist.Load(str)
					size := len(list) / 2
					for i := 0; i < size; i++ {
						logrus.Infoln("TypeHashZipList", "field=>value", key, list[2*i], list[2*i+1])
					}
				case rdb.TypeListZipList:
					list := ziplist.Load(str)
					for i := 0; i < len(list); i++ {
						logrus.Infoln("TypeListZipList", key, list[i])
					}
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
