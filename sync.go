package xxcache

import (
	"github.com/sirupsen/logrus"
	"time"
)

// Sync this will clear the local cache,so you should call this at first
func (cache *Cache)Watch(repl Replication) {
	var ack = time.NewTicker(time.Second)
	for {
		select {
		case err := <-repl.err:
			logrus.Warnln("replication", err)
		case msg := <-repl.messages:
			logrus.Infoln(msg)
		case <-ack.C:
			if err := repl.Ack(); err != nil {
				logrus.Warnln("replication Ack", err)
			}
		}
	}
}
