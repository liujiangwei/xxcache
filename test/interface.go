package main

import "github.com/sirupsen/logrus"

type I interface {

}

func main() {
	var i I
	i = &I{}
	switch i.(type) {
	case nil:
		logrus.Infoln("nil")
	}
}
