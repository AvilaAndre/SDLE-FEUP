package main

import (
	"log"
	"fmt"
	"os"
	// "bufio"
	"github.com/zeromq/goczmq"
	utils "sdle.com/mod/utils"
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
	
	utils.SendMessage(orchestrator, fmt.Sprintf("new_connection %s", "127.0.0.1"))

	ack := utils.ReceiveMessage(orchestrator)

	fmt.Println(string(ack[0]));
}
