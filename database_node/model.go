package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

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

var nodeJoin sync.Mutex

func nodeAction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/**
	 * Upon receiving this message, the node is in charge of propagating this message to every node and await for their response
	 */
	case http.MethodPut:
		{
			nodeJoin.Lock()

			if !startedSolo && len(nodes.getNodes()) == 0 {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Attempted to join node with no cluster"))
				log.Println("Rejected new join")
				nodeJoin.Unlock()
				return
			}

			target := make(map[string]string)

			json.NewDecoder(r.Body).Decode(&target)

			newNode := node{address: target["address"], port: target["port"]}

			if nodes.hasNode(newNode) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("There is already a node with that data"))
				nodeJoin.Unlock()
				return
			}

			var successfulPropagation bool = true
			propagatedChan := make(chan bool)

			knownNodes := nodes.getNodes()

			// propagates the message that a node was just added to the cluster
			for i := 0; i < len(knownNodes); i++ {
				// Use go channels to perform this requests and await for the responses
				go propagateAdd(knownNodes[i], target, propagatedChan)
			}

			// wait for the responses
			for i := 0; i < len(knownNodes); i++ {
				var result bool = <-propagatedChan
				successfulPropagation = successfulPropagation && result
			}

			// TODO: If the node already exists then the response should be to accept it
			if successfulPropagation && nodes.addNode(newNode) {
				// This node should inform the new node of all nodes existnt on the network
				nodesData := make(map[string][]map[string]string)

				nodesData["nodes"] = append(nodesData["nodes"], map[string]string{"address": serverHostname, "port": serverPort})
				for i := 0; i < len(knownNodes); i++ {
					nodesData["nodes"] = append(nodesData["nodes"], map[string]string{"address": knownNodes[i].address, "port": knownNodes[i].port})
				}

				jsonData, err := json.Marshal(nodesData)
				if err != nil {
					log.Fatalf("Error happened in JSON marshal. Err: %s", err)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				w.Write(jsonData)
			} else {
				log.Println("Node ADD Rejected")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("There is already a node with that data"))
			}
			nodeJoin.Unlock()

		}
	/**
	 * Upon receiving this message, the node knows a new connection is being propagated, accepting with 200 or rejecting with 400
	 */
	case http.MethodPost:
		{
			nodeJoin.Lock()
			target := make(map[string]string)

			json.NewDecoder(r.Body).Decode(&target)

			newNode := node{address: target["address"], port: target["port"]}

			// TODO: should return true
			if nodes.hasNode(newNode) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("There is already a node with that data"))
				nodeJoin.Unlock()
				return
			}

			nodes.addNode(newNode)

			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Node added successfully"))
			nodeJoin.Unlock()
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
	// TODO: Add a timeout
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
		fmt.Println(nodeToSend.address, nodeToSend.port, "rejected the new node")
	}
}
