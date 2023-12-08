package main

import (
	"encoding/json"
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

func nodeAdd(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/**
	 * Upon receiving this message, the node stores the new node in his hash ring
	 */
	case http.MethodPut:
		{
			target := make(map[string]string)

			decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

			if !decoded {
				return
			}

			var isServer bool = target["address"] == serverHostname && target["port"] == serverPort

			// Adds the new node to the cluster
			ring.AddNode(target["address"], target["port"], isServer)

			nodesOnTheRing := ring.GetNodes()

			nodesData := make(map[string][]map[string]string)
			nodesData["nodes"] = make([]map[string]string, len(nodesOnTheRing))

			for _, value := range ring.GetNodes() {
				nodesData["nodes"] = append(nodesData["nodes"], map[string]string{"address": value.Address, "port": value.Port})
			}

			jsonData, err := json.Marshal(nodesData)

			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			w.Write(jsonData)

		}
	default:
		{
			protocol.WrongRequestType(w)
		}
	}
}

func handleGossip(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/**
	 * Upon receiving this message, the node is in charge of propagating this message to every node and await for their response
	 */
	case http.MethodPost:
		target := make(map[string][]map[string]string)

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}
		

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("This node is operating normally"))
	}
}

// This responds and deals with the first pull-push request for anti-entropy mechanism
func handleGossipAntiEntropyRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/**
	 * Upon receiving this message, the node needs to answer the incoming gossip anti-entropy pull
	 with the lists he have in common with the sender and have a different hash for the awsets for those lists
	 */
	case http.MethodPost:
		target := make(map[string][]map[string]string)

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}
		
		//



		w.WriteHeader(http.StatusOK)
		w.Write([]byte("This node is operating normally"))
		w.Write([]byte("This node will answer the anti-entropy pull"))
	}
}
//This deals with the second phase of the pull-push anti-entropy mechanism
func handleAntiEntropyResponse(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/**
	 * Upon receiving this message, the node needs  to send the asked shoppingLists and merge the common ShoppingLists sended in gossipAntiEntropy first request
	 */
	case http.MethodPut:
		target := make(map[string][]map[string]string)

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}

		ring.CheckForNewNodes(target["nodes"], serverHostname, serverPort)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("This node is operating normally"))
		w.Write([]byte("This node is will answer the anti-entropy pull"))
	}
}
