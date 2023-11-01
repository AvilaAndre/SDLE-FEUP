package main

import (
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
	
	utils.SendMessage(new_connection_socket, protocol.ConnectMessage(own_endpoint.String(), dataPort))

	ack := utils.ReceiveMessageTimeout(new_connection_socket, 2000)

	if (ack == nil) {
		return false
	} else {
		return true
	}
}