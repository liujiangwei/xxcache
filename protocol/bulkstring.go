package protocol

import "strconv"

type BulkStringMessage struct {
	length int
	eof string
	Data string
}

func(message BulkStringMessage)String()string{
	return message.Data
}

func(message BulkStringMessage)Serialize()string{
	var str string
	if len(message.Data) == 0{
		if message.length < 0{
			// nil
			str += string(BulkString) + "-1\r\n"
		}else{
			str += string(BulkString) + "0\r\n"
		}
	}else{
		str = string(BulkString) + strconv.Itoa(len(message.Data)) + "\r\n"
		str += message.Data + "\r\n"
	}

	return str
}

func NewBulkStringMessage(data string, eof string, isNil bool) *BulkStringMessage{
	 message := &BulkStringMessage{
	 	Data:data,
	}

	if len(eof) == 40{
		message.eof = eof
		message.length = len(eof)
	}

	if isNil{
		message.Data = ""
		message.length = -1
		message.eof = ""
	}

	return message
}