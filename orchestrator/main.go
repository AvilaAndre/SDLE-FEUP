package main

import (
	"log"
	// "bufio"
	"os"

	"github.com/zeromq/goczmq"
	utils "sdle.com/mod/utils"
)


var listenerPort = "6873"
var cluster DatabaseClusterOrganization = *NewDatabaseClusterOrganization()


func main() {
	// var passive_orchestrator int

	log.Println("Started Orchestrator from IP " + utils.GetOutboundIP().String())

	argsWithoutProg := os.Args[1:];

	if len(argsWithoutProg) < 1 {
		log.Fatal("A port must be specified to initialize an orchestrator.")
	}

	listenerPort = argsWithoutProg[0];

	new_connection_listener := goczmq.NewSock(goczmq.Rep)
	defer new_connection_listener.Destroy()

	_, r1 := new_connection_listener.Bind("tcp://*:" + listenerPort)

	if r1 != nil {
		log.Fatal(r1)
	}

	poller, err := goczmq.NewPoller(new_connection_listener)

	if (err != nil) {
		log.Fatal(err)
	}
	
	for {
		u := poller.Wait(-1)
		
		switch u {
			// Listens for new connections
		case new_connection_listener:
			var new_db_node *DatabaseNode = HandleNewConnection(u)
			
			// Add new DatabaseNode to the respective array
			cluster.AddNewDatabaseNode(new_db_node)
		}
	}
}
