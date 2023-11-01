package protocol


const (
	ACKNOWLEDGE = "ACKNOWLEDGE"
	NEW_CONNECTION = "NEW_CONNECTION"
	CONNECTION_ACCEPTED = "CONNECTION_ACCEPTED"
	CONNECTION_REJECTED = "CONNECTION_REJECTED"
)

func AcknowledgeMessage() [][]byte {
	return message1(ACKNOWLEDGE)
}

func ConnectMessage(ip string, port string) [][]byte {
	return message3(NEW_CONNECTION, ip, port)
}

func AcceptConnectionMessage() [][]byte {
	return message1(CONNECTION_ACCEPTED)
}

func RejectConnectionMessage() [][]byte {
	return message1(CONNECTION_REJECTED)
}

func message3(header string, arg1 string, arg2 string) [][]byte {
	return [][]byte{[]byte(header), []byte(arg1), []byte(arg2)}
}

func message1(header string) [][]byte {
	return [][]byte{[]byte(header)}
}