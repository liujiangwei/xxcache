package protocol

import (
	"strconv"
	"strings"
)

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