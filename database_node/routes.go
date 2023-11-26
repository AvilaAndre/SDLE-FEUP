package main

import (
	"net/http"
)

func registerRoutes() {
	// http.HandleFunc("/", getRoot)
	http.HandleFunc("/list", handleList)
	http.HandleFunc("/gossip", handleGossip)
	http.HandleFunc("/node/add", nodeAdd)
	http.HandleFunc("/ping", getPing)
}
