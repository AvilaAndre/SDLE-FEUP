package main

import (
	"encoding/json"
	"log"
	"net/http"
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
	 * Upon receiving this message, the node is in charge of propagating this message to every node and await for their response
	 */
	case http.MethodPut:
		{
			target := make(map[string]string)

			json.NewDecoder(r.Body).Decode(&target)

			if ring.addNode(target["address"], target["port"]) {
				nodesOnTheRing := ring.getNodes()

				nodesData := make(map[string][]map[string]string)
				nodesData["nodes"] = make([]map[string]string, len(nodesOnTheRing))

				for _, value := range ring.getNodes() {
					nodesData["nodes"] = append(nodesData["nodes"], map[string]string{"address": value.address, "port": value.port})
				}

				// If the add is successful, send accepted status and the nodes in the ring
				jsonData, err := json.Marshal(nodesData)

				if err != nil {
					log.Fatalf("Error happened in JSON marshal. Err: %s", err)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				w.Write(jsonData)
			} else {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("Attempted to join node with no cluster"))
				log.Println("Rejected new join")
			}

		}
	default:
		{
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Wrong request type"))
		}
	}
}
