package main

import (
	"log"
	"fmt"
	"os"
	// "bufio"
	"github.com/zeromq/goczmq"
)


func main() {
	argsWithoutProg := os.Args[1:];

	if len(argsWithoutProg) < 1 {
		log.Fatal("A port must be specified to initialize a database node.")
	}

	var orchestrator_port string = argsWithoutProg[0];

	fmt.Print("Hello World from a Database Node!\n");

	orchestrator := goczmq.NewSock(goczmq.Req)
	defer orchestrator.Destroy()

	r1 := orchestrator.Connect("tcp://localhost:" + orchestrator_port)

	if r1 != nil {
		log.Fatal(r1)
	}
	

	var msg []byte = []byte(fmt.Sprintf("new_connection %s", "127.0.0.1"))
	orchestrator.SendMessage([][]byte{msg})

	ack, err := orchestrator.RecvMessage()
	
	if (err != nil) {
		log.Fatal(err)
	}

	fmt.Println(string(ack[0]));
}
