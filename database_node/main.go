package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Server Running...")

    argsWithoutProg := os.Args[1:]
	var server_port string = argsWithoutProg[0];

	if server_port == "" {
		fmt.Printf("A server port must be specified") 
		os.Exit(1)
	}


	registerRoutes();

	err := http.ListenAndServe(fmt.Sprintf(":%s", server_port), nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err) 
		os.Exit(1)
	}
}