package redis

import (
	"errors"
	"fmt"
	"github.com/liujiangwei/xxcache/database"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"sync"
	"time"
)

type Replication struct {
	MasterAddr string

	Id     string
	Offset int

	Messages chan Message
	Err      chan error
	stop     chan struct{}

	Database *database.Database

	conn *Conn

	sync.Mutex
}

const DefaultReplicationMessages = 10000

func (repl *Replication) Stop() {
}

func (repl *Replication) SyncWithRedis() error {
	repl.Messages = make(chan Message, DefaultReplicationMessages)
	repl.Err = make(chan error, 1)

	if repl.MasterAddr == "" {
		return errors.New("redis master addr is required")
	}

	if conn, err := Connect(repl.MasterAddr); err != nil {
		err = errors.New("failed to connect to master," + err.Error())
		return err
	} else {
		repl.conn = conn
	}

	if err := repl.Ping(); err != nil {
		logrus.Warn("replication error", err)
		return err
	}

	if err := repl.Sync(); err != nil {
		logrus.Warnln("replication error", err)
		return err
	}

	// wait for message from master
	go repl.WaitForMessage()

	return nil
}

func (repl *Replication) WaitForMessage() {
	for {
		if message, err := repl.conn.Recv(); err != nil {
			close(repl.Messages)
			repl.Err <- err
			return
		} else {
			repl.Messages <- message
		}
	}
}

func (repl *Replication) Ack() error {
	cmd := StringCommand{
		BaseCommand: NewBaseCommand("ReplConf", "Ack", strconv.Itoa(repl.Offset)),
	}
	if err := repl.conn.Send(cmd.Serialize()); err != nil {
		err = errors.New("failed send Ack to master," + err.Error())
		return err
	}

	return nil
}

func (repl *Replication) Ping() error {
	cmd := StringCommand{
		BaseCommand: NewBaseCommand("Ping"),
	}
	if message, err := repl.conn.SendAndWaitReply(cmd.Serialize()); err != nil {
		return err
	} else if message.String() != PONG.String() {
		err = errors.New(fmt.Sprintf("Ping, receive %s from %s", message.String(), repl.MasterAddr))
		return err
	}

	return nil
}

const RdbFile = "./tmp.rdb"

func (repl *Replication) Sync() (err error) {
	cmd := StringCommand{
		BaseCommand: NewBaseCommand("SYNC"),
	}

	if err = repl.conn.Send(cmd.Serialize()); err != nil {
		err = errors.New("Sync error," + err.Error())
		return err
	}

	var fp *os.File
	if fp, err = os.OpenFile(RdbFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm); err != nil {
		err = errors.New("failed to open rdb file," + err.Error())
		return err
	}

	if protocol, length, err := repl.conn.ReadWithWriter(fp); err != nil {
		err = errors.New("failed to save rdb file," + err.Error())
		return err
	} else {
		logrus.Infoln("redis rdb file", Protocol(protocol), length)
	}

	return nil
}

func (repl *Replication) Stat(duration time.Duration) {
	for range time.NewTicker(duration).C {
		logrus.Infoln(fmt.Sprintf("Stat Id [%s] Offset[%d]", repl.Id, repl.Offset))
	}
}

func (repl *Replication) SetReplicationId(id string) {
	repl.Id = id
}

func (repl *Replication) SetReplicationOffset(offset int) {
	repl.Offset = offset
}