package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"sdle.com/mod/protocol"
)

func getPing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "pong"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}

	w.Write(jsonResp)
}

//https://www.google.com/search?q=md5+hashing&oq=md5+hashing&gs_lcrp=EgZjaHJvbWUyCQgAEEUYORiABDIHCAEQABiABDIHCAIQABiABDIHCAMQABiABDIICAQQABgWGB4yCAgFEAAYFhgeMggIBhAAGBYYHjIICAcQABgWGB4yCAgIEAAYFhge0gEIMjY3NGowajeoAgCwAgA&sourceid=chrome&ie=UTF-8

func nodeAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/**
	 * Upon receiving this message, the node is in charge of propagating this message to every node and await for their response
	 */
	case http.MethodPut:
		{
			target := make(map[string]string)

			json.NewDecoder(r.Body).Decode(&target)

			knownNodes := nodes.getNodes()

			newNode := node{address: target["address"], port: target["port"]}
			if nodes.hasNode(newNode) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("There is already a node with that data"))
				return
			}

			var successfulPropagation bool = true
			for i := 0; i < len(nodes.getNodes()); i++ {
				fmt.Println("propagating...")
				// Use go channels to perform this requests and await for the responses
				protocol.SendRequestWithData(http.MethodPost, knownNodes[i].address, knownNodes[i].port, "/node/add", target)
			}

			if successfulPropagation && nodes.addNode(newNode) {
				log.Println("propagation successful", nodes.getNodes())
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("Node added successfully"))
			} else {
				log.Println("propagation failed")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("There is already a node with that data"))
				return
			}
		}
	/**
	 * Upon receiving this message, the node knows a new connection is being propagated, accepting with 200 or rejecting with 400
	 */
	case http.MethodPost:
		{
			target := make(map[string]string)

			json.NewDecoder(r.Body).Decode(&target)

			newNode := node{address: target["address"], port: target["port"]}

			if nodes.hasNode(newNode) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("There is already a node with that data"))
				return
			}

			nodes.addNode(newNode)
			log.Println("New node added", newNode)

			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Node added successfully"))
		}
	/**
	 * TODO: Delete node from cluster
	 */
	case http.MethodDelete:
		{
			// TODO: plan delete
		}
	default:
		{
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong request type"))
		}
	}
}
