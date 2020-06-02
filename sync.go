package xxcache

import (
	"github.com/sirupsen/logrus"
	"time"
)

// Sync this will clear the local cache,so you should call this at first
func (cache *Cache) SyncWithRedis() (err error) {
	var repl = Replication{
		masterAddr: cache.option.Addr,
		stat:       true,
	}

	if err := repl.Start(); err != nil {
		logrus.Warnln(err)
		return err
	}
	defer repl.Stop()

	var ack = time.NewTicker(time.Second)
	for {
		select {
		case err = <-repl.err:
			return err
		case msg := <-repl.messages:
			logrus.Infoln(msg)
		case <-ack.C:
			if err := repl.ack(); err != nil {
				return err
			}
		}
	}

	return nil
}
