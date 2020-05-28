package xxcache
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
