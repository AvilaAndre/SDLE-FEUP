package main

import (
	"log"

	"github.com/zeromq/goczmq"
	protocol "sdle.com/mod/protocol"
	utils "sdle.com/mod/utils"
)

func HandleNewConnection(socket *goczmq.Sock) *DatabaseNode {
	// Receive new connection message
	var msg [][]byte = utils.ReceiveMessage(socket)

	// Connection messages have three parts
	if (len(msg) < 3) {
		// Reject the new connection
		utils.SendMessage(socket, protocol.RejectConnectionMessage())
		return nil
	}

	if (string(msg[0]) != protocol.NEW_CONNECTION) {
		log.Printf("Wrong Header")
		// Reject the new connection
		utils.SendMessage(socket, protocol.RejectConnectionMessage())
		return nil
	}

	// Log the message received
	log.Printf(string(msg[0]) + " " + string(msg[1]) + ":" + string(msg[2]))

	var endpoint string = string(msg[1])
	var port string = string(msg[2])

	// Accept new Database Node Connection
	utils.SendMessage(socket, protocol.AcceptConnectionMessage())

	// Create Socket to be able to message Database Node
	db_node_socket := goczmq.NewSock(goczmq.Req)

	// Connect to the Socket
	if !(utils.ConnectSocketTimeout(db_node_socket, endpoint, port, 2000)){
		// Reject the new connection because no connection could be made
		log.Println(protocol.NEW_CONNECTION, "rejected, unable to connect back.")
		utils.SendMessage(socket, protocol.RejectConnectionMessage())
		return nil
	}

	return NewDatabaseNode(endpoint, port, db_node_socket)
}
