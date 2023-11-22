package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"sdle.com/mod/protocol"
	"sdle.com/mod/utils"
)

func startServerAndJoinCluster(serverPort string, loadBalancerAddress string, loadBalancerPort string, ownData map[string]string) {
	serverRunning := make(chan bool)

	go startServer(serverPort, serverRunning)

	// FIXME: Should I wait for the server to start?

	go joinCluster(loadBalancerAddress, loadBalancerPort, ownData)

	<-serverRunning // waits for the server to close
}

func startServer(serverPort string, serverRunning chan bool) {
	err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed")
	} else if err != nil {
		log.Printf("error starting server: %s\n", err)
	}

	serverRunning <- true
}

func joinCluster(loadBalancerAddress string, loadBalancerPort string, ownData map[string]string) {
	if loadBalancerAddress != "" && loadBalancerPort != "" {
		nodes.addNode(node{address: loadBalancerAddress, port: loadBalancerPort})
		r, err := protocol.SendRequestWithData(http.MethodPut, loadBalancerAddress, loadBalancerPort, "/node/add", ownData)
		utils.CheckErr(err)

		log.Println("Tried to join the cluster:", r.Status)
	}

}
