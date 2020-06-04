package main

import (
	"bytes"
	"encoding/binary"
	"github.com/go-redis/redis/v7"
	"github.com/liujiangwei/xxcache/redis/zipmap"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {

	logrus.Fatalln(time.Now().UnixNano() / 1000000)

	zm := zipmap.New()
	log.Println(zm.Set("abc", "v"))
	log.Fatalln(zm)

	b := bytes.NewBuffer([]byte{})
	if err := binary.Write(b, binary.LittleEndian, int32(123456)); err != nil{
		log.Fatalln(err)
	}else{
		log.Fatalln(b.Bytes())
	}


	// *lenptr = ((buf[0]&0x3F)<<8)|buf[1]
	buf := []byte{115, 101}
	// 01110011 01100101   00111111
	logrus.Fatal((buf[0] & 0x3F) << 8)

	listener, err := net.Listen("tcp", "localhost:10000")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(conn net.Conn) {
			for {
				var data = make([]byte, 10)
				length, err := conn.Read(data)
				log.Println(conn.RemoteAddr(), length, err, string(data))
				time.Sleep(1)
			}
		}(conn)
	}
}


func testData() {
	client := redis.NewClient(&redis.Options{})
	logrus.Fatalln(client.Get("a").Result())
	for i := 0; i < 1; i++ {
		// string
		client.Set(strconv.Itoa(i), i, 0).Result()
		client.Set("string" + strconv.Itoa(i), i, 0)

		//list
		//client.Del("list")
		client.LPush("list", i)
		//client.Del("list-string")
		client.LPush("list-string", "string-" + strconv.Itoa(i)+"-a")

		// [29 0 0 0 19 0 0 0 2 0 0 7 102 105 101 108 100 45 48 9 7 118 97 108 117 101 45 48 255]
		//set
		//client.Del("set")
		client.SAdd("set", i)
		client.SAdd("set-100", i* 100)
		//client.Del("set-100")
		client.SAdd("set-10000", i * 10000)
		//client.Del("set-1000000")
		client.SAdd("set-1000000", i * 1000000)

		//sorted set
		//client.Del("zset")
		client.ZAdd("zset", &redis.Z{
			Score:  float64(i),
			Member:  i+1 ,
		})

		// hash
		client.HSet("hash", i, i+1)
		client.HSet("hash-string", "field-" + strconv.Itoa(i), "value-" + strconv.Itoa(i))
	}
}
