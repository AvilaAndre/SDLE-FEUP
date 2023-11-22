package main

import "net/http"

func registerRoutes() {
	// http.HandleFunc("/", getRoot)
	// http.HandleFunc("/nodes", getNodes)
	// http.HandleFunc("/add", getAdd)
	http.HandleFunc("/node/add", nodeAction)
	http.HandleFunc("/ping", getPing)
}
