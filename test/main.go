package main

import (
	"log"
)

type Greeting func(name string)

func (g Greeting) English(name string){
	g("hello, " + name)
}

func (g Greeting) Chinese(name string){
	g("你好，" + name)
}

type Person struct {
	Greeting
}

func main() {
	p := new(Person)
	p.Greeting = func(name string) {
		log.Println("before")

		log.Println(name)

		log.Println("after")
	}

	p.Chinese("小王")
	p.English("jack")
}