package protocol

type ErrorMessage struct {
	Data string
}

func(message ErrorMessage)String()string{
	return message.Data
}

func(message ErrorMessage)Serialize()string{
	return string(Error) + message.Data + "\r\n"
}
