package main

import (
	"fmt"
	"log"
	"os"

	// "bufio"
	"github.com/zeromq/goczmq"
	"sdle.com/mod/protocol"
	utils "sdle.com/mod/utils"
)

var own_endpoint = utils.GetOutboundIP()

var orchestrator_connect_endpoint = "localhost"
var orchestrator_connect_port = "6873"

var data_port = "6875"
var data_socket *goczmq.Sock = goczmq.NewSock(goczmq.Rep)

var own_id int = -1
var left_neighbour_socket *goczmq.Sock = nil
var right_neighbour_socket *goczmq.Sock = nil

func main() {
	defer data_socket.Destroy()

	argsWithoutProg := os.Args[1:];

	if len(argsWithoutProg) < 3 {
		log.Fatal("A port must be specified to initialize a database node.")
	}

	orchestrator_connect_endpoint = argsWithoutProg[0];
	orchestrator_connect_port = argsWithoutProg[1];
	data_port = argsWithoutProg[2];

	fmt.Print("Hello World from a Database Node!\n");

	_, r1 := data_socket.Bind("tcp://*:" + data_port)
	
	if r1 != nil {
		log.Fatal(r1)
	}

	if ConnectToOrchestrator(orchestrator_connect_endpoint, orchestrator_connect_port) {
		log.Println("Connected to orchestrator sucessfully")
	} else {
		log.Println("Failed to connect to the orchestrator")
		return
	}

	poller, err := goczmq.NewPoller(data_socket)

	if (err != nil) {
		log.Fatal(err)
	}
	
	for {
		u := poller.Wait(-1)
		
		switch u {
			// Listens for new connections
		case data_socket:
			var msg [][]byte = utils.ReceiveMessage(u)

			// Messages should have a header
			if (len(msg) < 1) {
				utils.SendMessage(u, protocol.DenyMessage())
			}

			HandleNewMessage(u, msg)
		}
	}
}

