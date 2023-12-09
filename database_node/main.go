package main

import (
	"fmt"
	"log"
	"os"
	hash_ring "sdle.com/mod/hash_ring"
	"sdle.com/mod/utils"
)

var serverPort string = ""
var serverHostname string = ""

var ring hash_ring.HashRing

var database DatabaseInstance

const replicationFactor int = 5

func main() {
	

	var loadBalancerAddress string = ""
	var loadBalancerPort string = ""

	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		serverPort = argsWithoutProg[0]
		serverHostname = utils.GetOutboundIP().String()
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
	log.Printf("Node starting... %s:%s", serverHostname, serverPort)

	ring.Initialize()
	database.initialize(serverHostname, serverPort)

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
