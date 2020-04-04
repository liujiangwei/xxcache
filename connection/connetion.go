package connection

import (
	"bufio"
	"github.com/liujiangwei/xxcache/protocol"
	"net"
	"time"
)

type Connection struct {
	Conn           net.Conn
	lastActiveTime time.Time
	reader         *bufio.Reader
	writer         *bufio.Writer
}

func (connection *Connection) Wait() (protocol.Message, error) {
	defer func() {
		connection.lastActiveTime = time.Now()
	}()

	return protocol.ReadOne(connection.reader)
}

func (connection *Connection) Reply(message protocol.Message) error{
	defer func() {
		connection.lastActiveTime = time.Now()
	}()

	_, err := connection.writer.Write([]byte(message.Serialize()))
	connection.writer.Flush()

	return err
}

func(connection *Connection)Close() error{
	return connection.Conn.Close()
}

func New(conn net.Conn) *Connection {
	return &Connection{
		Conn:           conn,
		lastActiveTime: time.Now(),
		reader:         bufio.NewReader(conn),
		writer:bufio.NewWriter(conn),
	}
}

func Connect(address string) (*Connection, error) {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return nil, err
	}

	return New(conn), nil
}

func listern(address string) {

}
