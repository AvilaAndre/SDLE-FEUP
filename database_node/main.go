package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"sdle.com/mod/protocol"
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

	ownData := make(map[string]string)

	ownData["address"] = utils.GetOutboundIP().String()
	ownData["port"] = serverPort

	if loadBalancerAddress != "" && loadBalancerPort != "" {
		nodes.addNode(node{address: loadBalancerAddress, port: loadBalancerPort})
		protocol.SendRequestWithData(http.MethodPut, loadBalancerAddress, loadBalancerPort, "/node/add", ownData)
	}

	err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
