package xxcache

import (
	"context"
	"github.com/cornelk/hashmap"
	"github.com/liujiangwei/xxcache/redis"
	"github.com/sirupsen/logrus"
	"sync"
)

type Cache struct {
	// todo connection pool
	ConnPool *Pool

	option Option

	database []*Database
	cdb      *Database

	//serverInfo map[string]string

	replicationId           string
	replicationId2          string
	replicationOffset       int
	secondReplicationOffset int

	lock sync.Mutex
}

type Option struct {
	Addr string
}


// Sync this will clear the local cache,so you should call this at first
func (cache *Cache) Sync() error{
	cache.lock.Lock()
	defer cache.lock.Unlock()

	// redis info to set database
	infoRaw, err := cache.Info()
	if err != nil {
		return  err
	}
	info, err := redis.ParseInfo(infoRaw)
	if err != nil {
		return err
	}

	cache.replicationId = info.Replication.MasterReplicationId
	cache.replicationId2 = info.Replication.MasterReplicationId2
	cache.replicationOffset = info.Replication.MasterReplOffset
	cache.secondReplicationOffset = info.Replication.SecondReplOffset

	logrus.Infoln(cache.ConfigGet("*"))

	// initialize database for try psync first
	//conn , err := redis.Connect(cache.option.Addr)
	//if err != nil{
	//	logrus.Fatal(err)
	//}
	//
	//if err := conn.Send(redis.ConvertToMessage("PING")); err != nil{
	//	logrus.Fatal("failed to ping to master", err)
	//}
	//
	//if msg , err := conn.Recv();err != nil{
	//	logrus.Fatal("SYNC", "failed to receive from master", err)
	//}else{
	//	logrus.Infoln("SYNC","receive pong from master", msg.String())
	//}
	//
	//var cmd redis.Message
	//// SYNC_CMD_WRITE,conn,"REPLCONF", //                "listening-port",portstr, NULL
	//cmd = redis.ConvertToMessage("REPLCONF", "listening-port", "6380")
	//if err := conn.Send(cmd); err != nil{
	//	logrus.Fatalln("SYNC", "failed to send listen port to master", err)
	//}
	//
	//if message, err := conn.Recv(); err != nil{
	//	logrus.Fatalln("SYNC", "send listen port to master", err)
	//}else{
	//	logrus.Infoln("SYNC", "send listen port to master", message.String())
	//}
	//
	//if err := conn.Send(redis.ConvertToMessage("REPLCONF", "ip-address", "127.0.0.1")); err != nil{
	//	logrus.Fatalln("SYNC", "send ip-address to master", err)
	//}
	//
	//if msg, err := conn.Recv(); err != nil{
	//	logrus.Fatalln("SYNC","send ip-address to master", err)
	//}else{
	//	logrus.Infoln("SYNC", "send ip-address to master", msg.String())
	//}
	//
	//logrus.Infoln("try psync", cache.replicationId, cache.replicationOffset)
	//if err := conn.Send(redis.ConvertToMessage("PSYNC", cache.replicationId, strconv.Itoa(cache.replicationOffset))); err != nil{
	//	logrus.Fatalln("SYNC", "rsync from master", err)
	//}
	//
	//if message, err := conn.Recv(); err != nil{
	//	logrus.Fatalln("SYNC", "receive PSYNC", err)
	//}else{
	//	if strings.ToUpper(message.String()) == "CONTINUE"{
	//		logrus.Infoln("SYNC", "receive PSYNC", message.String())
	//	}else{
	//		logrus.Warnln("failed to psync from redis master, receive", message.String())
	//	}
	//}
	//
	//logrus.Println(conn.ReadMessage())
	//
	//if protocol, err := conn.ReadProtocol(); err != nil || protocol != redis.ProtocolBulkString{
	//	logrus.Fatalf("failed to read protocol from message [%s]", protocol)
	//}
	//
	//if err := conn.DiscardEof(); err != nil{
	//	logrus.Fatalf("failed to read protocol from message  [%s]", err)
	//}
	//
	//var version = make([]byte, 9)
	//if _, err := conn.Reader.Read(version); err != nil{
	//	logrus.Fatalln(err)
	//}

	return nil
}

func (cache *Cache)initializeDatabase(num int64)  {
	for i:= int64(0); i< num; i++{
		cache.database[i] = &Database{dict: hashmap.HashMap{}}
	}
}

func New(option Option) (*Cache , error){
	cache := Cache{
		option:option,
	}

	if pool, err := initPool(4, option.Addr); err != nil{
		return nil, err
	}else{
		cache.ConnPool = pool
	}


	return &cache, nil
}

func (cache Cache) Process(command Command) error {
	return cache.ConnPool.Exec(context.Background(), command)
}

func (cache Cache) Ping() (string, error){
	cmd := NewStringCommand("Ping")
	cmd.err = cache.Process(&cmd)

	return cmd.val, cmd.err
}

func (cache Cache) Info(sections ...string) (string, error){
	args := append([]string{"INFO"}, sections...)

	cmd := NewStringCommand(args...)
	cmd.err = cache.Process(&cmd)

	return cmd.Result(), cmd.err
}

func (cache Cache) ConfigGet(sections ...string) (map[string]string, error){
	args := append([]string{"CONFIG", "GET"}, sections...)

	cmd := NewStringStringCommand(args...)
	cmd.err = cache.Process(&cmd)

	return cmd.Result(), cmd.err
}