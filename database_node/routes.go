package main

import (
	"net/http"
)

func registerRoutes() {
	// http.HandleFunc("/", getRoot)
	
	http.HandleFunc("/operation", handleOperation)
	http.HandleFunc("/list", handleCoordenator)
	http.HandleFunc("/gossip", handleGossip)
	http.HandleFunc("/gossip/antiEntropy/request", handleGossipAntiEntropyRequest)
	http.HandleFunc("/gossip/antiEntropy/response", handleAntiEntropyResponse)
	http.HandleFunc("/gossip/antiEntropy/finalStep", handleAntiEntropyFinalStep)
	http.HandleFunc("/node/add", nodeAdd)
	http.HandleFunc("/ping", getPing)
}
