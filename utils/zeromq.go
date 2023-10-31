package utils

import (
	"log"
	"github.com/zeromq/goczmq"
	"net"
)

func ConnectSocketTimeout(socket *goczmq.Sock, endpoint string, port string, timeout int) bool {
	socket.SetConnectTimeout(timeout)
	err := socket.Connect("tcp://" + endpoint + ":" + port)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}


func ReceiveMessage(socket *goczmq.Sock) [][]byte {
	socket.SetRcvtimeo(-1) // Won't Timeout

	msg, err := socket.RecvMessage()
		
	if (err != nil) {
		log.Println(err)
		return nil
	}

	return msg
}

func ReceiveMessageTimeout(socket *goczmq.Sock, timeout int) [][]byte {
	socket.SetRcvtimeo(timeout) // Waits {timeout} milliseconds before timing out 

	msg, err := socket.RecvMessage()
		
	if (err != nil) {
		log.Println(err, "Message receive time out")
		return nil
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

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}