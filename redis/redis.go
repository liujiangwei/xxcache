package redis

func ConvertToMessage(args ...string) Message {
	msg := ArrayMessage{}
	for _, arg := range args {
		msg = append(msg, NewBulkStringMessage(arg))
	}
	return msg
}
