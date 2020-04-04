package protocol

type SimpleStringMessage struct {
	Data string
}

func(message SimpleStringMessage)String()string{
	return message.Data
}

func(message SimpleStringMessage)Serialize()string{
	return string(SimpleString) + message.Data + "\r\n"
}
