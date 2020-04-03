package service

import (
	"github.com/liujiangwei/xxcache/connection"
	"github.com/liujiangwei/xxcache/protocol"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type Server struct {
	Commander
	listener                net.Listener
	master                  *connection.Connection
	masterReplicationId     string
	masterReplicationOffset int
	pingMasterTimeInterval  time.Duration
	logger                  *logrus.Logger
}

func (server *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	server.listener = listener
	server.logger = logrus.New()

	// init service
	server.init()

	// start to receive connection
	server.accept()

	return nil
}

func (server *Server) init() {

}

func (server *Server) accept() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			server.logger.Println("accept new connect error", err)
			_ = conn.Close()
		}

		server.logger.Info("new connection")
		go server.handle(connection.New(conn))
	}
}

func (server *Server) handle(conn *connection.Connection) {
	for {
		msg , err := conn.Wait()
		if err != nil {
			server.logger.Info(err)
		}

		handler, args := convertToHandler(msg)
		if handler != nil{
			conn.Reply(handler.Exec(server, args))
		}else{
			conn.Reply(ErrCommanderNotFoundMessage)
		}
	}
}

var ErrCommanderNotFoundMessage = protocol.SimpleStringMessage{Data: "command not found"}

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

func (server *Server) Get(key string) string{
	return key
}

func (server *Server)Set(key string, value string) string{
	return "OK"
}