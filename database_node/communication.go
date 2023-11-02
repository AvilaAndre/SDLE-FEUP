package main

import (
	"log"
	"strconv"

	"github.com/zeromq/goczmq"
	protocol "sdle.com/mod/protocol"
	utils "sdle.com/mod/utils"
)

func ConnectToOrchestrator(orchestrator_endpoint string, orchestrator_port string) bool {

	new_connection_socket := goczmq.NewSock(goczmq.Req)
	defer new_connection_socket.Destroy()

	if (!utils.ConnectSocketTimeout(new_connection_socket, orchestrator_endpoint, orchestrator_port, 1000)) {
		return false
	}
	
	utils.SendMessage(new_connection_socket, protocol.ConnectMessage(own_endpoint.String(), data_port))

	ack := utils.ReceiveMessageTimeout(new_connection_socket, 2000)

	if (ack == nil) {
		return false
	} else {
		return true
	}
}

func HandleNewMessage(socket *goczmq.Sock, msg [][]byte) {
	switch string(msg[0]) {
	case protocol.UPDATE_NODE_ID:
		new_id, er := strconv.Atoi(string(msg[1]))
		if er != nil {
			log.Println("Failed to update ID", er)
			utils.SendMessage(socket, protocol.AcknowledgeMessage())
			break
		}
		own_id = new_id
		log.Println("Updated ID to ", own_id)
		utils.SendMessage(socket, protocol.AcknowledgeMessage())
	case protocol.UPDATE_NODE_LEFT_NEIGHBOUR:
		var endpoint string = string(msg[1])
		var port string = string(msg[2])
		
		left_neighbour_socket = goczmq.NewSock(goczmq.Req)
		if (utils.ConnectSocketTimeout(left_neighbour_socket, endpoint, port, 2000)) {
			log.Println("Updated left neighbour.")
			utils.SendMessage(socket, protocol.AcknowledgeMessage())
		} else {
			log.Println("Failed to update left neighbour.")
			utils.SendMessage(socket, protocol.DenyMessage())
		}
	case protocol.UPDATE_NODE_RIGHT_NEIGHBOUR:
		var endpoint string = string(msg[1])
		var port string = string(msg[2])
		
		right_neighbour_socket = goczmq.NewSock(goczmq.Req)
		if (utils.ConnectSocketTimeout(right_neighbour_socket, endpoint, port, 2000)) {
			log.Println("Updated right neighbour.")
			utils.SendMessage(socket, protocol.AcknowledgeMessage())
		} else {
			log.Println("Failed to update right neighbour.")
			utils.SendMessage(socket, protocol.DenyMessage())
		}
	default:
		log.Println("Received unknown header", string(msg[0]))
		utils.SendMessage(socket, protocol.DenyMessage())
	}
}