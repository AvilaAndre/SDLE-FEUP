package main

import (
	"fmt"
	"log"
	"os"

	// "bufio"
	"github.com/zeromq/goczmq"
	utils "sdle.com/mod/utils"
)

var own_endpoint = utils.GetOutboundIP()

var orchestrator_connect_endpoint = "localhost"
var orchestrator_connect_port = "6873"

var dataPort = "6875"
var dataSocket *goczmq.Sock = goczmq.NewSock(goczmq.Rep)

func main() {
	defer dataSocket.Destroy()

	argsWithoutProg := os.Args[1:];

	if len(argsWithoutProg) < 3 {
		log.Fatal("A port must be specified to initialize a database node.")
	}

	orchestrator_connect_endpoint = argsWithoutProg[0];
	orchestrator_connect_port = argsWithoutProg[1];
	dataPort = argsWithoutProg[2];

	fmt.Print("Hello World from a Database Node!\n");

	_, r1 := dataSocket.Bind("tcp://*:" + dataPort)
	
	if r1 != nil {
		log.Fatal(r1)
	}

	if ConnectToOrchestrator(orchestrator_connect_endpoint, orchestrator_connect_port) {
		log.Println("Connected to orchestrator sucessfully")
	} else {
		log.Println("Failed to connect to the orchestrator")
		return
	}

	poller, err := goczmq.NewPoller(dataSocket)

	if (err != nil) {
		log.Fatal(err)
	}
	
	for {
		u := poller.Wait(-1)
		
		switch u {
			// Listens for new connections
		case dataSocket:
			var msg [][]byte = utils.ReceiveMessage(u)
			log.Println(string(msg[0]))
		}
	}
}
