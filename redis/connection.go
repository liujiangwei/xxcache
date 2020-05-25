package redis

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
)

type Connection struct {
	conn   net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

func (c *Connection) Recv() (Message, error) {
	return c.ReadMessage()
}

func (c *Connection) Send(message Message) error {
	_, err := c.Writer.Write([]byte(message.Serialize()))
	c.Writer.Flush()

	return err
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn:   conn,
		Reader: bufio.NewReader(conn),
		Writer: bufio.NewWriter(conn),
	}
}

func Connect(address string) (*Connection, error) {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return nil, err
	}

	return NewConn(conn), nil
}

func (c Connection) RecvToFile(file string) (int64, error) {
	fp, err := os.OpenFile(file, os.O_CREATE, os.ModeAppend)
	if err != nil {
		return 0, err
	}

	return c.Reader.WriteTo(fp)
}

// protocol length
func (c Connection) ReadProtocol() (protocol, error) {
	var p byte
	p, err := c.Reader.ReadByte()
	return protocol(p), err
}

func (c Connection) readBulkString(length int) string {
	if length > 0 {
		var data = make([]byte, length)
		_, err := c.Reader.Read(data)
		if err != nil {

		}
		_, _, _ = c.Reader.ReadLine()

		return string(data)
	}

	return ""
}

func (c Connection) DiscardEof() error {
	_, _, err := c.Reader.ReadLine()

	return err
}

func (c Connection) ReadMessage() (Message, error) {
	protocol, err := c.ReadProtocol()
	if err != nil {
		return nil, err
	}

	switch protocol {
	case ProtocolBulkString:
		// $1
		bs, _, err := c.Reader.ReadLine()
		if err != nil {
			return nil, err
		}
		// get length
		length, err := strconv.Atoi(string(bs))
		if err != nil {
			return nil, err
		}

		// read data
		var data = make([]byte, length)
		if _, err := c.Reader.Read(data); err != nil {
			return nil, err
		}

		if err := c.DiscardEof(); err != nil {
			return nil, err
		}

		msg := NewBulkStringMessage(string(data))

		return msg, nil
	case ProtocolSimpleString:
		line, _, err := c.Reader.ReadLine()
		if err != nil {
			return nil, err
		}

		return SimpleStringMessage(string(line)), nil
	case ProtocolError:
		line, _, err := c.Reader.ReadLine()
		if err != nil {
			return nil, err
		}
		return ErrorMessage(string(line)), nil

	case ProtocolInt:
		line, _, err := c.Reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if number, err := strconv.Atoi(string(line)); err != nil {
			return nil, err
		} else {
			return IntMessage(number), nil
		}
	case ProtocolArray:
		line, _, err := c.Reader.ReadLine()
		if err != nil {
			return nil, err
		}
		number, err := strconv.Atoi(string(line))
		if err != nil {
			return nil, err
		}

		var messages ArrayMessage
		for i := 0; i < number; i++ {
			if message, err := c.ReadMessage(); err != nil {
				return nil, err
			} else {
				messages = append(messages, message)
			}
		}

		return messages, nil
	default:
		return nil, errors.New(fmt.Sprintf("[%s]", string(protocol)))
	}
}
