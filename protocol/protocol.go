package protocol

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type RedisProtocolReader struct {
	reader *bufio.Reader
}

type Protocol string

//字符串：以 $ 开始
const BulkString = Protocol("$")
// 简单字符串：以"+" 开始， 如："+OK\r\n"
const SimpleString = Protocol("+")
//整数：以":"开始，如：":1\r\n"
const Int = Protocol(":")
//数组：以 * 开始
const Array = Protocol("*")
//错误：以"-" 开始，如："-ERR Invalid Synatx\r\n"
const Error = Protocol("-")


type Message interface {
	String() string
	Serialize() string
}

type BulkStringMessage struct {
	Data string
}

func(message BulkStringMessage)String()string{
	return message.Data
}

func(message BulkStringMessage)Serialize()string{
	str := string(BulkString) + strconv.Itoa(len(message.Data)) + "\r\n"
	str += message.Data + "\r\n"

	return str
}


type SimpleStringMessage struct {
	Data string
}

func(message SimpleStringMessage)String()string{
	return message.Data
}

func(message SimpleStringMessage)Serialize()string{
	return string(SimpleString) + message.Data + "\r\n"
}

type IntMessage struct {
	Data int
}

func(message IntMessage)String()string{
	return strconv.Itoa(message.Data)
}
func(message IntMessage)Serialize()string{
	return string(Int) + strconv.Itoa(message.Data) + "\r\n"
}

type ArrayMessage struct {
	Data []Message
}

func(message ArrayMessage)String()string{
	var str []string

	for _, m := range message.Data{
		str = append(str, m.String())
	}

	return strings.Join(str, "")
}

func(message ArrayMessage)Serialize()string{
	str := string(Array) + strconv.Itoa(len(message.Data)) + "\r\n"

	for _, m := range message.Data{
		str += m.Serialize()
	}

	return str
}

type ErrorMessage struct {
	Data string
}

func(message ErrorMessage)String()string{
	return message.Data
}

func(message ErrorMessage)Serialize()string{
	return string(Error) + message.Data + "\r\n"
}

func ReadOne(reader *bufio.Reader) (Message, error){
	var protocol = make([]byte, 1)
	// 读取协议
	if _, err := reader.Read(protocol); err != nil {
		return nil, err
	}
	switch Protocol(string(protocol)) {
	case BulkString:
		if line, _, err := reader.ReadLine(); err != nil {
			return nil, err
		} else {
			if length, err := strconv.Atoi(string(line)); err != nil {
				// $EOF
				var message = bytes.Buffer{}
				for {
					if data, _,err := reader.ReadLine(); err != nil{
						return nil, err
					}else{
						if string(data) == string(line){
							break
						}else{
							message.Write(data)
						}
					}
				}
				return BulkStringMessage{Data: message.String()}, nil
			} else {
				// $120
				var data = make([]byte, length)
				_,_ = reader.Read(data)

				_, _ , _ = reader.ReadLine()

				return BulkStringMessage{
					Data: string(data),
				}, nil
			}
		}
	case SimpleString:
		if line, _, err := reader.ReadLine(); err != nil {
			return nil, err
		} else {
			return SimpleStringMessage{Data: string(line)},nil
		}
	case Error:
		if line, _, err := reader.ReadLine(); err != nil {
			return nil, err
		} else {
			return ErrorMessage{Data: string(line)},nil
		}
	case Int:
		if line, _, err := reader.ReadLine(); err != nil {
			return nil, err
		} else if line, err := strconv.Atoi(string(line)); err != nil{
			return nil, err
		}else{
			return IntMessage{Data: line}, nil
		}
	case Array:
		if rows, _, err := reader.ReadLine(); err != nil {
			return nil,err
		} else if rows, err := strconv.Atoi(string(rows)); err != nil {
			return nil, err
		} else {
			var arr ArrayMessage
			for i := 0; i < rows; i++ {
				if message, err := ReadOne(reader); err != nil{
					return nil, err
				}else{
					arr.Data = append(arr.Data, message)
				}
			}
			return arr, nil
		}
	default:
		if line, _, err  := reader.ReadLine(); err != nil{
			return nil, errors.New(fmt.Sprintf("[%s] [%s] [%s]", string(protocol), string(line), err.Error()))
		}else{
			return nil, errors.New(fmt.Sprintf("[%s] [%s]", string(protocol), string(line)))
		}
	}
}

var ReplyPong = SimpleStringMessage{Data:"PONG"}

var ReplyOk = SimpleStringMessage{Data:"OK"}
