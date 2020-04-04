package protocol

import "strconv"

type IntMessage struct {
	Data int
}

func(message IntMessage)String()string{
	return strconv.Itoa(message.Data)
}
func(message IntMessage)Serialize()string{
	return string(Int) + strconv.Itoa(message.Data) + "\r\n"
}