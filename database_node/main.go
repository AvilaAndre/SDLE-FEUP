package main

import (
	"log"
	"fmt"
	"os"
	// "bufio"
)


func main() {
	argsWithoutProg := os.Args[1:];

	if len(argsWithoutProg) < 1 {
		log.Fatal("A port must be specified to initialize a database node.")
	}

	var orchestrator_endpoint string = argsWithoutProg[0];
	var orchestrator_port string = argsWithoutProg[1];

	fmt.Print("Hello World from a Database Node!\n");

	if ConnectToOrchestrator(orchestrator_endpoint, orchestrator_port) {
		log.Println("Connected to orchestrator sucessfully")
	} else {
		log.Println("Failed to connect to the orchestrator")
	}
}
