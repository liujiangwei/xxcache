package main

import (
	"bytes"
	"encoding/binary"
	"github.com/liujiangwei/xxcache/redis/zipmap"
	"github.com/sirupsen/logrus"
	"log"
	"net"
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