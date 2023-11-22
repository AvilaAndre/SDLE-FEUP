package main

import (
	"fmt"
	"os"

	"sdle.com/mod/utils"
)

var nodes nodeManager

func main() {
	fmt.Println("Server Running...")

	var serverPort string = ""
	var loadBalancerAddress string = ""
	var loadBalancerPort string = ""

	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		serverPort = argsWithoutProg[0]
	}

	if len(argsWithoutProg) == 3 {
		loadBalancerAddress = argsWithoutProg[1]
		loadBalancerPort = argsWithoutProg[2]
	}

	if serverPort == "" {
		fmt.Println("A server port must be specified")
		os.Exit(1)
	}

	registerRoutes()

	if loadBalancerAddress != "" && loadBalancerPort != "" {
		ownData := make(map[string]string)

		ownData["address"] = utils.GetOutboundIP().String()
		ownData["port"] = serverPort
		startServerAndJoinCluster(serverPort, loadBalancerAddress, loadBalancerPort, ownData)
	} else {
		serverRunning := make(chan bool)
		startServer(serverPort, serverRunning)
		<-serverRunning // waits for the server to close
	}
}
