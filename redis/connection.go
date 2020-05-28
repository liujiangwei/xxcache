package redis

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

type Conn struct {
	conn   net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

func (c *Conn) Recv() (Message, error) {
	return c.readMessage()
}

func (c *Conn) send(message Message) error {
	_, err := c.Writer.Write([]byte(message.Serialize()))

	if err != nil{
		return  err
	}

	return c.Writer.Flush()
}

func (c *Conn) Send(message Message) (Message, error){
	if err := c.send(message); err != nil{
		return nil, err
	}

	return c.Recv()
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

// Protocol length
func (c Conn) readProtocol() (Protocol, error) {
	var p byte
	p, err := c.Reader.ReadByte()
	return Protocol(p), err
}

func (c Conn) readBulkString(length int) string {
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

func (c Conn) discardEof() error {
	_, _, err := c.Reader.ReadLine()

	return err
}

// save message content to io reader
func (c Conn) ReadWithWriter(writer io.Writer) (Protocol, int, error){
	protocol, err := c.readProtocol()
	if err != nil{
		return protocol, 0, err
	}

	buf := bufio.NewWriter(writer)

	defer func() {
		_ = buf.Flush()
	}()

	switch protocol {
	case ProtocolBulkString:
		bs, _, err := c.Reader.ReadLine()
		if err != nil {
			return ProtocolBulkString, 0, err
		}

		// get length
		length, err := strconv.Atoi(string(bs))
		if err != nil {
			return ProtocolBulkString, 0, err
		}

		for i :=0; i< length; i++{
			b, err := c.Reader.ReadByte()
			if err != nil{
				return ProtocolBulkString, 0,  err
			}
			if err := buf.WriteByte(b); err != nil{
				return ProtocolBulkString, 0, err
			}
		}

		return ProtocolBulkString, length, nil
	case ProtocolSimpleString:
		line, _, err := c.Reader.ReadLine()
		if err != nil {
			return ProtocolSimpleString, 0, err
		}
		n, err := buf.Write(line)

		return ProtocolSimpleString, n, err
	case ProtocolError:
		line, _, err := c.Reader.ReadLine()
		if err != nil {
			return ProtocolError, 0, err
		}
		n, err := buf.Write(line)
		return ProtocolError, n, err
	case ProtocolInt:
		line, _, err := c.Reader.ReadLine()
		if err != nil {
			return ProtocolInt, 0, err
		}
		n, err := buf.Write(line)
		return ProtocolInt, n, err
	case ProtocolArray:
		line, _, err := c.Reader.ReadLine()
		if err != nil {
			return ProtocolArray, 0 , err
		}

		number, err := strconv.Atoi(string(line))
		return ProtocolArray, number, err
	default:
		return protocol, 0, errors.New("unknown protocol")
	}
}

func (c Conn) readMessage() (Message, error) {
	protocol, err := c.readProtocol()
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

		if err := c.discardEof(); err != nil {
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
			if message, err := c.readMessage(); err != nil {
				return nil, err
			} else {
				messages = append(messages, message)
			}
		}

		return messages, nil
	default:
		return nil, errors.New("unknown protocol " + fmt.Sprintf("[%s]", string(protocol)))
	}
}



func NewConn(conn net.Conn) *Conn {
	return &Conn{
		conn:   conn,
		Reader: bufio.NewReader(conn),
		Writer: bufio.NewWriter(conn),
	}
}

func Connect(address string) (*Conn, error) {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		return nil, err
	}

	return NewConn(conn), nil
}