package protocol

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

type RedisProtocolReader struct {
	reader *bufio.Reader
}

type Protocol string

//字符串：以 $ 开始  $-1\r\n
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

type MessageProtocol struct {
	protocol Protocol
	length int
	data []byte
	eof string
	error error
}

func readProtocol(reader *bufio.Reader) MessageProtocol{
	var protocol = make([]byte, 1)

	var mp =MessageProtocol{}
	if _, err := reader.Read(protocol); err != nil {
		mp.error = err
		return mp
	}

	mp.protocol = Protocol(protocol)

	// for length or eof
	data, _, err  := reader.ReadLine()
	mp.data = data

	if err != nil{
		mp.error = err
		return mp
	}

	// eof
	if len(mp.data) == 40{
		mp.eof = string(mp.data)
	}else if length, err := strconv.Atoi(string(mp.data)); err != nil{
		mp.error = err
	}else{
		mp.length = length
	}

	return mp
}

func readBulkString(reader *bufio.Reader, protocol MessageProtocol) string{
	if protocol.eof != ""{
		var message = bytes.Buffer{}
		for {
			data, _,err := reader.ReadLine()

			if err != nil || string(data) == string(protocol.protocol){
				break
			}

			message.Write(data)
		}

		return message.String()
	}

	if protocol.length > 0{
		var data = make([]byte, protocol.length)
		_,_ = reader.Read(data)
		_, _ , _ = reader.ReadLine()

		return string(data)
	}

	return ""
}


func ReadOne(reader *bufio.Reader) (Message, error){
	mp := readProtocol(reader)
	switch mp.protocol{
	case BulkString:
		return BulkStringMessage{
			eof: mp.eof,
			Data:readBulkString(reader, mp),
		}, nil

	case SimpleString:
		return SimpleStringMessage{
			Data:string(mp.data),
		}, nil
	case Error:
		return ErrorMessage{Data:string(mp.data)}, nil
	case Int:
		if number, err := strconv.Atoi(string(mp.data)); err != nil{
			return nil, err
		}else{
			return IntMessage{Data: number}, nil
		}
	case Array:
		var arr ArrayMessage
		for i := 0; i < mp.length; i++ {
			if message, err := ReadOne(reader); err != nil{
				return nil, err
			}else{
				arr.Data = append(arr.Data, message)
			}
		}

		return arr, nil
	default:
		return nil, errors.New(fmt.Sprintf("[%s] [%s] [%s]", string(mp.protocol), string(mp.data), mp.error))
	}
}

var Nil = BulkStringMessage{
	Data: "",
	length: -1,
}

var OK = SimpleStringMessage{
	Data: "OK",
}