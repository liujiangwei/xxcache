package redis

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/liujiangwei/xxcache/redis/intset"
	"github.com/liujiangwei/xxcache/redis/rdb"
	"github.com/liujiangwei/xxcache/redis/ziplist"
	"github.com/liujiangwei/xxcache/redis/zipmap"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Replication struct {
	MasterAddr string
	RdbFile    string
	Id         string
	Offset     int
	database   int
	Messages   chan ReplicationMessage
	Replies    chan redis.Message
	Err        chan error
	stop       chan struct{}

	lastPing time.Time

	conn *Conn
	sync.Mutex
}

const DefaultReplicationMessages = 10000

func (repl *Replication) Stop() {
}

func (repl *Replication) SyncWithRedis() error {
	if repl.RdbFile == "" {
		return errors.New("rdb file is empty")
	}

	repl.Messages = make(chan ReplicationMessage, DefaultReplicationMessages)
	repl.Err = make(chan error, 1)

	if repl.MasterAddr == "" {
		return errors.New("redis master addr is required")
	}

	if conn, err := Connect(repl.MasterAddr); err != nil {
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
		return err
	}

	return nil
}

func (repl *Replication) WaitForMessage() {
	for {
		message, err := repl.conn.Recv()
		if err != nil {
			repl.Err <- err
			close(repl.Messages)
			return
		}

		if strings.HasPrefix(strings.ToLower(message.String()), "ping") {
			repl.lastPing = time.Now()
			continue
		}

		repl.Messages <- ReplicationMessage{
			Database: 0,
			Message:  message,
			Offset:   0,
		}
	}
}

func (repl *Replication) ack() error {
	cmd := StringCommand{
		BaseCommand: NewBaseCommand("ReplConf", "ack", strconv.Itoa(repl.Offset)),
	}

	logrus.Debugln("ack", strconv.Itoa(repl.Offset))
	if err := repl.conn.Send(cmd.Serialize()); err != nil {
		err = errors.New("failed send ack to master," + err.Error())
		return err
	}

	return nil
}

func (repl *Replication) Ack() {
	for range time.NewTicker(time.Second).C {
		if err := repl.ack(); err != nil {
			repl.Err <- err
		}
	}
}

func (repl *Replication) ping() error {
	cmd := StringCommand{
		BaseCommand: NewBaseCommand("ping"),
	}
	if message, err := repl.conn.SendAndWaitReply(cmd.Serialize()); err != nil {
		return err
	} else if message.String() != PONG.String() {
		err = errors.New(fmt.Sprintf("ping, receive %s from %s", message.String(), repl.MasterAddr))
		return err
	}

	return nil
}

func (repl *Replication) sync() (err error) {
	cmd := StringCommand{
		BaseCommand: NewBaseCommand("SYNC"),
	}

	if err = repl.conn.Send(cmd.Serialize()); err != nil {
		return errors.New("sync error," + err.Error())
	}

	var fp *os.File
	if fp, err = os.OpenFile(repl.RdbFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm); err != nil {
		return errors.New("failed to open rdb file," + err.Error())
	}

	if protocol, length, err := repl.conn.ReadWithWriter(fp); err != nil {
		return errors.New("failed to save rdb file," + err.Error())
	} else {
		logrus.Infoln("redis rdb file", protocol, length, repl.RdbFile)
	}

	return err
}

//psync from redis master
func (repl *Replication) pSync() (err error) {

	return err
}

func (repl *Replication) SetReplicationId(id string) {
	repl.Id = id
}

func (repl *Replication) SetReplicationOffset(offset int) {
	repl.Offset = offset
}

func (repl *Replication) Load() (err error) {
	if repl.RdbFile == "" {
		return errors.New("rdb file is empty")
	}

	var file *os.File
	if file, err = os.OpenFile(repl.RdbFile, os.O_RDONLY, os.ModePerm); err != nil {
		return errors.New("failed to open rdb file," + err.Error())
	}

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

		switch opCode {
		case rdb.OpCodeExpireTime:
			// 4 byte
			var expireS = make([]byte, 4)
			if _, err = io.ReadFull(buf, expireS); err != nil {
				return err
			}
			expiresTime = uint64(binary.LittleEndian.Uint32(expireS) * 1000)
		case rdb.OpCodeExpireTimeMs:
			var milliTime = make([]byte, 8)
			if _, err = io.ReadFull(buf, milliTime); err != nil {
				return err
			}
			expiresTime = binary.LittleEndian.Uint64(milliTime)
		case rdb.OpCodeFreq:
			var lfu byte
			if lfu, err = buf.ReadByte(); err != nil {
				return err
			}
			logrus.Warnln("lfu", lfu)
		case rdb.OpCodeIdle:
			var lru uint64
			if lru, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			logrus.Println("OpCodeIdle", lru)
		case rdb.OpCodeEof:
			logrus.Infoln("rdb done!!!")
			return nil
		case rdb.OpCodeSelectDB:
			if index, _, err := rdb.LoadLen(buf); err != nil {
				return errors.New("OpCodeSelectDB error, " + err.Error())
			} else {
				repl.database = int(index)
				logrus.Debugln("OpCodeSelectDB", index)
			}
		case rdb.OpCodeResizeDB:
			var dbSize, expiresSize uint64
			if dbSize, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			if expiresSize, _, err = rdb.LoadLen(buf); err != nil {
				return err
			}
			logrus.Debugln("OpCodeResizeDB ignored", dbSize, expiresSize)
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
				} else {
					cmd := NewBaseCommand("set", key, value)
					repl.Messages <- ReplicationMessage{
						Database: repl.database,
						Message:  cmd.Serialize(),
						Offset:   0,
					}
				}
			case rdb.TypeList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeList length", err)
					return err
				}

				var value string
				var args = []string{"lPush", key}
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeList value", err)
						return err
					} else {
						args = append(args, value)
					}
				}

				cmd := NewBaseCommand(args...)
				repl.Messages <- ReplicationMessage{
					Database: repl.database,
					Message:  cmd.Serialize(),
					Offset:   0,
				}

			case rdb.TypeSet:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeSet length", err)
					return err
				}
				var args = []string{"sAdd", key}

				for ; length > 0; length-- {
					if value, err := rdb.LoadString(buf); err != nil {
						return err
					} else {
						args = append(args, value)
					}
				}
				cmd := NewBaseCommand(args...)
				repl.Messages <- ReplicationMessage{
					Database: repl.database,
					Message:  cmd.Serialize(),
					Offset:   0,
				}
			case rdb.TypeZSet, rdb.TypeZSet2:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeZSet, TypeZSet2 length", err)
					return err
				}

				var args = []string{"zAdd", key}

				var value string
				var score float64
				for ; length > 0; length-- {
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeZSet, TypeZSet2 value", err)
						return err
					}
					args = append(args, value)
					if opCode == rdb.TypeZSet2 {
						if score, err = rdb.LoadBinaryDouble(buf); err != nil {
							logrus.Warnln("failed to load TypeZSet, TypeZSet2 value TypeZSet2", err)
							return err
						}
					} else {
						if score, err = rdb.LoadDouble(buf); err != nil {
							logrus.Warnln("failed to load TypeZSet, TypeZSet2 value TypeZSet2", err)
							return err
						}
					}

					args = append(args, strconv.FormatFloat(score, 'f', -1, 64))
				}

				cmd := NewBaseCommand(args...)
				repl.Messages <- ReplicationMessage{
					Database: repl.database,
					Message:  cmd.Serialize(),
					Offset:   0,
				}
			case rdb.TypeHash:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeHash length", err)
					return err
				}

				var args = []string{"zAdd", key}
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
					args = append(args, field, value)
				}

				cmd := NewBaseCommand(args...)
				repl.Messages <- ReplicationMessage{
					Database: repl.database,
					Message:  cmd.Serialize(),
					Offset:   0,
				}
			case rdb.TypeListQuickList:
				var length uint64
				if length, _, err = rdb.LoadLen(buf); err != nil {
					logrus.Warnln("failed to load TypeListQuickList length", err)
					return err
				}
				var args = []string{"rPush", key}
				for ; length > 0; length-- {
					var value string
					if value, err = rdb.LoadString(buf); err != nil {
						logrus.Warnln("failed to load TypeListQuickList value", err)
						return err
					}
					args = append(args, ziplist.Load(value)...)
				}

				cmd := NewBaseCommand(args...)
				repl.Messages <- ReplicationMessage{
					Database: repl.database,
					Message:  cmd.Serialize(),
					Offset:   0,
				}
			case rdb.TypeHashZipMap, rdb.TypeListZipList, rdb.TypeSetIntSet, rdb.TypeZSetZipList, rdb.TypeHashZipList:
				var str string
				if str, err = rdb.LoadString(buf); err != nil {
					logrus.Warnln("failed to load TypeHashZipMap... length", key, err)
					return err
				}

				switch opCode {
				case rdb.TypeHashZipMap:
					var args = []string{"hmSet", key}
					hash := zipmap.Load(str)
					for field, value := range hash {
						args = append(args, field, value)
					}

					cmd := NewBaseCommand(args...)
					repl.Messages <- ReplicationMessage{
						Database: repl.database,
						Message:  cmd.Serialize(),
						Offset:   0,
					}
				case rdb.TypeSetIntSet:
					var args = []string{"sAdd", key}
					set := intset.Load(str)
					for _, v := range set {
						args = append(args, strconv.Itoa(int(v)))
					}

					cmd := NewBaseCommand(args...)
					repl.Messages <- ReplicationMessage{
						Database: repl.database,
						Message:  cmd.Serialize(),
						Offset:   0,
					}
				case rdb.TypeZSetZipList:
					var args = []string{"zAdd", key}
					list := ziplist.Load(str)
					for i := 0; i < len(list)/2; i++ {
						args = append(args, list[2*i], list[2*i+1])
					}
					cmd := NewBaseCommand(args...)
					repl.Messages <- ReplicationMessage{
						Database: repl.database,
						Message:  cmd.Serialize(),
						Offset:   0,
					}
				case rdb.TypeHashZipList:
					var args = []string{"hmSet", key}
					list := ziplist.Load(str)
					args = append(args, list...)
					cmd := NewBaseCommand(args...)
					repl.Messages <- ReplicationMessage{
						Database: repl.database,
						Message:  cmd.Serialize(),
						Offset:   0,
					}
				case rdb.TypeListZipList:
					var args = []string{"rPush", key}
					list := ziplist.Load(str)
					args = append(args, list...)
					cmd := NewBaseCommand(args...)
					repl.Messages <- ReplicationMessage{
						Database: repl.database,
						Message:  cmd.Serialize(),
						Offset:   0,
					}
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
				args := []string{"pExpire", key, strconv.Itoa(int(expiresTime))}
				cmd := NewBaseCommand(args...)
				repl.Messages <- ReplicationMessage{
					Database: repl.database,
					Message:  cmd.Serialize(),
					Offset:   0,
				}
				expiresTime = 0
			}
		}
	}

	return nil
}

type ReplicationMessage struct {
	Database int
	Message  Message
	Offset   int
}
