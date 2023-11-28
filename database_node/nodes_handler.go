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

			if !protocol.DecodeRequestBody(w, r.Body, target) {
				return
			}

			// Adds the new node to the cluster
			ring.addNode(target["address"], target["port"])

			nodesOnTheRing := ring.getNodes()

			nodesData := make(map[string][]map[string]string)
			nodesData["nodes"] = make([]map[string]string, len(nodesOnTheRing))

			for _, value := range ring.getNodes() {
				nodesData["nodes"] = append(nodesData["nodes"], map[string]string{"address": value.address, "port": value.port})
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

		if !protocol.DecodeRequestBody(w, r.Body, target) {
			return
		}

		ring.checkForNewNodes(target["nodes"])

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("This node is operating normally"))
	}
}