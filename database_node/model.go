package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

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

var propagateLock sync.Mutex

func nodeAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/**
	 * Upon receiving this message, the node is in charge of propagating this message to every node and await for their response
	 */
	case http.MethodPut:
		{
			log.Println("Received PUT")
			target := make(map[string]string)

			json.NewDecoder(r.Body).Decode(&target)

			newNode := node{address: target["address"], port: target["port"]}
			if nodes.hasNode(newNode) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("There is already a node with that data"))
				return
			}

			var successfulPropagation bool = true
			propagatedChan := make(chan bool)

			propagateLock.Lock()
			knownNodes := nodes.getNodes()

			log.Println("acquired propagate lock", knownNodes)

			// propagates the message that a node was just added to the cluster
			for i := 0; i < len(knownNodes); i++ {
				fmt.Println("propagating...")
				// Use go channels to perform this requests and await for the responses
				go propagateAdd(knownNodes[i], target, propagatedChan)
			}

			// wait for the responses
			for i := 0; i < len(knownNodes); i++ {
				log.Println(i, len(knownNodes))
				var result bool = <-propagatedChan
				log.Println(i, result)
				successfulPropagation = successfulPropagation && result
			}
			propagateLock.Unlock()
			log.Println("freed propagate lock")

			if successfulPropagation && nodes.addNode(newNode) {
				log.Println("propagation successful", nodes.getNodes())
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte("Node added successfully"))
			} else {
				log.Println("propagation failed")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("There is already a node with that data"))
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

			fmt.Println("zzzzzz")
			time.Sleep(10 * time.Second)
			fmt.Println("?")

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

func propagateAdd(nodeToSend node, data map[string]string, propagatedChan chan bool) {
	r, err := protocol.SendRequestWithData(http.MethodPost, nodeToSend.address, nodeToSend.port, "/node/add", data)

	if err != nil {
		propagatedChan <- false
		return
	}

	// 202 means accepted, therefore, the node add was accepted
	if r.StatusCode == 202 {
		propagatedChan <- true
	} else {
		propagatedChan <- false
	}
}
