package main

import (
	"log"
	"net"
	"time"
)

func main() {
	conn , err := net.Dial("tcp", "localhost:10000")
	if err != nil{
		log.Fatal(err)
	}

	for {
		length ,err := conn.Write([]byte("a"))
		log.Println(length, err)

		time.Sleep(time.Millisecond * 1000)
	}
}
