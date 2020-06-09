package main

import "github.com/sirupsen/logrus"

type A struct {
	B
}
func(a *A)test(){
	logrus.Infoln("a")
}

type B struct {

}

func(b *B)test(){
	logrus.Infoln("b")
}
func main() {
	a := A{}
	a.test()
}
