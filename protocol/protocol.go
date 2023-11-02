package protocol


const (
	ACKNOWLEDGE = "ACKNOWLEDGE"
	DENY = "DENY"
	NEW_CONNECTION = "NEW_CONNECTION"
	CONNECTION_ACCEPTED = "CONNECTION_ACCEPTED"
	CONNECTION_REJECTED = "CONNECTION_REJECTED"
	UPDATE_NODE_ID = "UPDATE_NODE_ID"
	UPDATE_NODE_LEFT_NEIGHBOUR = "UPDATE_NODE_LEFT_NEIGHBOUR"
	UPDATE_NODE_RIGHT_NEIGHBOUR = "UPDATE_NODE_RIGHT_NEIGHBOUR"
)

func message1(header string) [][]byte {
	return [][]byte{[]byte(header)}
}

func message2(header string, arg1 string) [][]byte {
	return [][]byte{[]byte(header), []byte(arg1)}
}

func message3(header string, arg1 string, arg2 string) [][]byte {
	return [][]byte{[]byte(header), []byte(arg1), []byte(arg2)}
}

func AcknowledgeMessage() [][]byte {
	return message1(ACKNOWLEDGE)
}

func DenyMessage() [][]byte {
	return message1(DENY)
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

func UpdateNodeIDMessage(id string) [][]byte {
	return message2(UPDATE_NODE_ID, id)
}

func UpdateNodeLeftNeighbourMessage(endpoint string, port string) [][]byte {
	return message3(UPDATE_NODE_LEFT_NEIGHBOUR, endpoint, port)
}

func UpdateNodeRightNeighbourMessage(endpoint string, port string) [][]byte {
	return message3(UPDATE_NODE_RIGHT_NEIGHBOUR, endpoint, port)
}