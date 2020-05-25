package main

import (
	"log"
	"sync"
	"time"
)

func main() {

	var sm = sync.Map{}
	go func() {
		sm.Store("test", "test_test_test_test_test_test_test_test_test_test_test_test")

	}()

	for i:=0; i<1000; i ++{
		log.Println(sm.Load("test"))

		time.Sleep(time.Microsecond * 1000)
	}
}
