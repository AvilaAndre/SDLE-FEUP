package main

import "net/http"


func registerRoutes() {
	// http.HandleFunc("/", getRoot)
	// http.HandleFunc("/nodes", getNodes)
	// http.HandleFunc("/add", getAdd)
	http.HandleFunc("/ping", getPing)
}