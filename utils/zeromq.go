package utils

import (
	"log"
	"github.com/zeromq/goczmq"
)

func ReceiveMessage(socket *goczmq.Sock) [][]byte {
	msg, err := socket.RecvMessage()
		
	if (err != nil) {
		log.Fatal(err)
	}

	return msg
}

func SendMessage(socket *goczmq.Sock, message_string string) {
	var message []byte = []byte(message_string)
	err := socket.SendMessage([][]byte{message})
	if (err != nil) {
		log.Fatal(err)
	}
}