package service

import (
	"github.com/liujiangwei/xxcache/command"
	"github.com/liujiangwei/xxcache/connection"
	"github.com/liujiangwei/xxcache/dict"
	"github.com/liujiangwei/xxcache/entry"
	"github.com/sirupsen/logrus"
	"net"
)

type Server struct {
	listener                net.Listener

	//db
	Database []*entry.Database
	DatabaseSelected *entry.Database

	// for replication
	master                  *connection.Connection
	masterReplicationId     string
	masterReplicationOffset int

	// server config
	Option  Option

	// log
	logger                  *logrus.Logger
}


func (server *Server) Start() error {
	server.logger = logrus.New()

	// init service
	server.init(newOption())

	// start to receive connection
	return nil
}

func (server *Server) Listen(address string) error{
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server.listener = listener

	return server.accept()
}

func (server *Server) init(option *Option) {
	if option.DatabaseSize <= 0{
		option.DatabaseSize = defaultOptionDatabaseSize
	}

	for i:=0; i< option.DatabaseSize; i++{
		server.Database = append(server.Database, &entry.Database{
			Dict:dict.Default(),
		})
	}

	server.DatabaseSelected = server.Database[0]
}


func (server *Server) accept() error{
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			server.logger.Println("accept connect error", err)
			_ = server.listener.Close()
			_ = conn.Close()
			return  err
		}

		server.logger.Info("new connection")
		go server.handle(connection.New(conn))
	}
}

func (server *Server) handle(conn *connection.Connection) {
	defer conn.Close()
	for {
		msg , err := conn.Wait()
		if err != nil {
			server.logger.Warn("read from client error, close the connection", err)
			break
		}

		commander := command.Convert(msg)

		if  err := conn.Reply(commander.Exec(server)); err != nil{
			server.logger.Warn("reply to client error, close the connection", err)
			break
		}
	}
}

func (server *Server) Sync(address string) error {
	if conn, err := connection.Connect(address); err != nil {
		return err
	} else {
		server.master = conn
	}

	return nil
}

func (server *Server) Ping(message string) string{
	if message == ""{
		message = "PONG"
	}

	return message
}

