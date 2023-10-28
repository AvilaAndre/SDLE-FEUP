package main

import (
	"log"
	"fmt"
	// "bufio"
	"os"
	"github.com/zeromq/goczmq"
)


func main() {

	argsWithoutProg := os.Args[1:];

	if len(argsWithoutProg) < 1 {
		log.Fatal("A port must be specified to initialize an orchestrator.")
	}

	var listenerPort string = argsWithoutProg[0];

	fmt.Print("Hello World from the Orchestrator!\n");


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
			msg, err := u.RecvMessage()
		
			if (err != nil) {
				log.Fatal(err)
			}

			log.Printf(string(msg[0]))

			var ack []byte = []byte("ACK")
			u.SendMessage([][]byte{ack})
			return
		}

	}
}
