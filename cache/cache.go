package cache

import (
	"github.com/liujiangwei/xxcache/command"
	"github.com/liujiangwei/xxcache/dict"
	"github.com/liujiangwei/xxcache/entry"
	"github.com/liujiangwei/xxcache/rconn"
	"github.com/liujiangwei/xxcache/rsync"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Cache struct {
	listener                net.Listener

	//db
	database         []*entry.Database
	databaseSelected *entry.Database

	// for replication
	master                  *rconn.Connection
	masterReplicationId     string
	masterReplicationOffset int
	masterReplicationState int

	// server config
	Option  Option

	// log
	logger                  *logrus.Logger

	// lock
	sync.RWMutex
}


func (cache *Cache) Start() error {
	cache.logger = logrus.New()
	// init cache
	cache.init(newOption())

	// start to receive rconn
	return nil
}

func (cache *Cache) Listen(address string) error{
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	cache.listener = listener

	return cache.accept()
}

func (cache *Cache) init(option *Option) {
	if option.DatabaseSize <= 0{
		option.DatabaseSize = defaultOptionDatabaseSize
	}

	for i:=0; i< option.DatabaseSize; i++{
		cache.database = append(cache.database, entry.NewDatabase(dict.Default()))
	}

	cache.databaseSelected = cache.database[0]
}


func (cache *Cache) accept() error{
	for {
		conn, err := cache.listener.Accept()
		if err != nil {
			cache.logger.Println("accept connect error", err)
			_ = cache.listener.Close()
			_ = conn.Close()
			return  err
		}

		cache.logger.Info("new rconn")
		go cache.handle(rconn.New(conn))
	}
}

func (cache *Cache) handle(conn *rconn.Connection) {
	defer func() {
		err := conn.Close()
		cache.logger.Warnln("rconn was closed because ", err)
	}()

	for {
		msg , err := conn.Recv()
		cache.logger.Infoln("receive message from client", msg)
		if err != nil {
			cache.logger.Warn("read from client error, close the rconn", err)
			break
		}

		cmd := command.Convert(msg)
		reply := cmd.Exec(cache)
		cache.logger.Infoln("start send message to client", msg)
		if  err := conn.Send(reply); err != nil{
			cache.logger.Warn("reply to client error, close the rconn", err, reply)
			break
		}
	}
}

func (cache *Cache) Sync(address string) error {
	master, err := rsync.NewRdbClient(address)
	if  err != nil {
		cache.logger.Warn("SYNC", "failed to connect to master", err)
		return err
	}
	if err := master.Send(command.ConvertToMessage("PING")); err != nil{
		cache.logger.Warn("SYNC", "failed to ping to master")
		return err
	}

	if msg , err := master.Recv();err != nil{
		cache.logger.Warn("SYNC", "failed to receive from master", err)
		return err
	}else{
		cache.logger.Infoln("SYNC","receive pong from master", msg.String())
	}

	var cmd rconn.Message
	// SYNC_CMD_WRITE,rconn,"REPLCONF", //                "listening-port",portstr, NULL
	cmd = command.ConvertToMessage("REPLCONF", "listening-port", "6380")
	if err := master.Send(cmd); err != nil{
		cache.logger.Warnln("SYNC", "failed to send listen port to master", err)

		return err
	}
	if message, err := master.Recv(); err != nil{
		cache.logger.Warnln("SYNC", "send listen port to master", err)

		return err
	}else{
		cache.logger.Infoln("SYNC", "send listen port to master", message.String())
	}

	cmd = command.ConvertToMessage("REPLCONF", "ip-address", "127.0.0.1")
	if err := master.Send(cmd); err != nil{
		cache.logger.Warnln("SYNC", "send ip-address to master", err)

		return err
	}

	if msg, err := master.Recv(); err != nil{
		cache.logger.Warnln("SYNC","send ip-address to master", err)

		return err
	}else{
		cache.logger.Infoln("SYNC", "send ip-address to master", msg.String())
	}

	cmd = command.ConvertToMessage("PSYNC", cache.masterReplicationId, strconv.Itoa(cache.masterReplicationOffset))
	if err := master.Send(cmd); err != nil{
		cache.logger.Warn("SYNC", "rsync from master", err)

		return err
	}
	if message, err := master.Recv(); err != nil{
		cache.logger.Warnln("SYNC", "receive PSYNC", err)

		return err
	}else{
		if strings.ToUpper(message.String()) == "CONTINUE"{

		}

		cache.logger.Infoln("SYNC", "receive PSYNC", message.String())
	}
	//
	//// try pSync
	//if err := cache.master.Send(cmd); err != nil{
	//	cache.logger.Warnln("faild to psync to master")
	//
	//	return nil
	//}
	//cmd = command.ConvertToMessage("SYNC")
	//if err := master.Send(cmd); err != nil{
	//	cache.logger.Warn("SYNC", "rsync from master", err)
	//
	//	return err
	//}

	if protocol, err := master.ReadProtocol(); err != nil || protocol != rconn.ProtocolBulkString{
		cache.logger.Fatalf("failed to read protocol from message [%s]", protocol)
	}

	if err := master.DiscardEof(); err != nil{
		cache.logger.Fatalf("failed to read protocol from message  [%s]", err)
	}

	var version = make([]byte, 9)
	if _, err := master.Reader.Read(version); err != nil{
		cache.logger.Fatal(err)
	}

	for{
		opCode,err := master.LoadOpCode()
		if err != nil{
			cache.logger.Fatal("opcode error", err)
		}

		switch opCode {
		case rsync.RdbOPCodeExpireTime:
			// 4 byte
			var time = make([]byte, 4)
			if _,err :=master.Reader.Read(time); err != nil{
				cache.logger.Fatal("RdbOPCodeExpireTime", err)
			}

			cache.logger.Println("RdbOPCodeExpireTime", time)

			continue
		case rsync.RdbOpCodeExpireTimeMs:
			var milliTime = make([]byte, 8)
			if _, err := master.Reader.Read(milliTime);err != nil{
				cache.logger.Fatal("RdbOpCodeExpireTime_MS")
			}

			cache.logger.Println("RdbOpCodeExpireTime_MS", string(milliTime))
			continue
		case rsync.RdbOpCodeFreq:
			var lfu byte
			lfu , err := master.Reader.ReadByte()
			if err != nil{
				cache.logger.Fatal("RdbOpCodeFreq", err)
			}

			cache.logger.Println("RdbOpCodeFreq", lfu)
			continue
		case rsync.RdbOpCodeIdle:
			lru, _, err := master.LoadLen()
			if err != nil{
				log.Fatal(err)
			}
			cache.logger.Println("RdbOpCodeIdle",lru)
			continue
		case rsync.RdbOpCodeEof:
			cache.logger.Println("RdbOpCodeEof")
			return nil
		case rsync.RdbOpCodeSelectDB:
			dbid,_, err := master.LoadLen()
			if err != nil{
				log.Fatal(err)
			}
			cache.logger.Println("RdbOpCodeSelectDB", dbid)
			continue
		case rsync.RdbOpCodeResizeDB:
			db_size , _, err :=  master.LoadLen()
			if err != nil{
				log.Fatal(err)
			}
			expires_size, _, err := master.LoadLen()
			if err != nil{
				log.Fatal(err)
			}
			cache.logger.Println("RdbOpCodeResizeDB", db_size, expires_size)
			continue
		case rsync.RdbOpCodeAux:
			k, err :=  master.LoadString()
			if err != nil{
				log.Fatal("RdbOpCodeAux key", err)
			}
			v, err := master.LoadString()
			if err != nil{
				log.Fatal("RdbOpCodeAux value", err)
			}

			if k.String() == "repl-id"{
				cache.masterReplicationId =  v.String()
			}else if k.String() == "repl-offset"{
				if offset, err := v.Int(); err != nil{
					cache.logger.Fatalf("failed to set repl-offset", v.String())
				}else{
					cache.masterReplicationOffset = offset
				}
			}

			cache.logger.Println("RdbOpCodeAux", k, v.Val())
			continue
		case rsync.RdbOpCodeModuleAux:
			log.Println("RdbOpCodeModuleAux")
			continue
		}

		k, err :=  master.LoadString()
		if err != nil{
			log.Fatal(err)
		}

		v, err := master.LoadObj(opCode)
		if err != nil{
			log.Fatal(err)
		}

		cache.logger.Println("key value", opCode, k.String(), v)

	}

	return nil
}

func (cache *Cache) syncReplicationOffset(){

}


func saveRdb(conn *rconn.Connection) error{
	if _, err := conn.RecvToFile("./tmp.rsync"); err !=nil{
		return err
	}


	return os.Rename("./tmp.rsync", "dump.rsync")
}

func loadRdb() {

}
