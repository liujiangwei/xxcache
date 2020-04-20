package rconn

import (
	"strconv"
	"strings"
)

type protocol byte

//字符串：以 $ 开始  $-1\r\n
const ProtocolBulkString = protocol('$')
// 简单字符串：以"+" 开始， 如："+OK\r\n"
const ProtocolSimpleString = protocol('+')
//整数：以":"开始，如：":1\r\n"
const ProtocolInt = protocol(':')
//数组：以 * 开始
const ProtocolArray = protocol('*')
//错误：以"-" 开始，如："-ERR Invalid Synatx\r\n"
const ProtocolError = protocol('-')

type Message interface {
	String() string
	Serialize() string
}

//type Protocol struct {
//	protocol protocol
//	length   int
//	data     []byte
//	eof      string
//	error    error
//}


var Nil = NewNilMessage()

var OK = SimpleStringMessage("OK")

// redis error message
type ArrayMessage []Message

func(messages ArrayMessage)String()string{
	var str []string

	for _, m := range messages{
		str = append(str, m.String())
	}

	return strings.Join(str, "")
}

func(messages ArrayMessage)Serialize()string{
	str := string(ProtocolArray) + strconv.Itoa(len(messages)) + MessageEOF

	for _, m := range messages{
		str += m.Serialize()
	}

	return str
}


const MessageEOF = "\r\n"
// redis bulk string message
type bulkStringMessage struct {
	string
	nil bool
}

func(message bulkStringMessage)String()string{
	return message.string
}

func(message bulkStringMessage)Serialize()string{
	var str = string(ProtocolBulkString)
	if message.nil{
		return str + "-1" + MessageEOF
	}

	str += strconv.Itoa(len(message.string)) + MessageEOF

	str += message.string + MessageEOF

	return str
}

func NewBulkStringMessage(str string) *bulkStringMessage {
	return &bulkStringMessage{
		string: str,
		nil:    false,
	}
}

func NewNilMessage() *bulkStringMessage{
	return &bulkStringMessage{
		string: "",
		nil:    true,
	}
}

// redis error message
type ErrorMessage string

func(message ErrorMessage)String()string{
	return string(message)
}

func(message ErrorMessage)Serialize()string{
	return string(ProtocolError) + string(message) + MessageEOF
}

// redis int message
type IntMessage int

func(message IntMessage)String()string{
	return strconv.Itoa(int(message))
}
func(message IntMessage)Serialize()string{
	return string(ProtocolInt) + strconv.Itoa(int(message)) + MessageEOF
}

type SimpleStringMessage string

func(message SimpleStringMessage)String()string{
	return string(message)
}

func(message SimpleStringMessage)Serialize()string{
	return string(ProtocolSimpleString) + string(message) + MessageEOF
}
