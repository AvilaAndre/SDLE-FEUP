package main

import (
	"fmt"
	"net"
	"github.com/zeromq/goczmq"
	utils "sdle.com/mod/utils"
)

func ConnectToOrchestrator(orchestrator_endpoint string, orchestrator_port string) bool {
	var own_endpoint net.IP = utils.GetOutboundIP()

	orchestrator := goczmq.NewSock(goczmq.Req)
	defer orchestrator.Destroy()

	if (!utils.ConnectSocketTimeout(orchestrator, orchestrator_endpoint, orchestrator_port, 1000)) {
		return false
	}
	
	utils.SendMessage(orchestrator, fmt.Sprintf("new_connection %s", own_endpoint))

	ack := utils.ReceiveMessageTimeout(orchestrator, 2000)

	if (ack == nil) {
		return false
	} else {
		fmt.Println(string(ack[0]));
		
		return true
	}
}