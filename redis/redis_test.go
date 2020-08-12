package redis

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestNew(t *testing.T) {
	client := NewClient(Option{
		Addr:         "127.0.0.1:6379",
		MaxRetry:     0,
		ReadTimeout:  0,
		WriteTimeout: 0,
		Timeout:      0,
		Database:     0,
	})

	logrus.Infoln(client.Get("a"))
}
