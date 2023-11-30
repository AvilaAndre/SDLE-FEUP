package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"sdle.com/mod/protocol"
)

func handleCoordenator(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodPost:
		target := make(map[string]string)

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}

		fmt.Println("COORDINATOR read", target)

		// The coordenator, upon receiving a read, reads locally and performs a read quorum
		// however, this coordenator may not be a holder of this information, in this case
		// it only performs the read quorum

		// FIXME: Can have multiple replicas in the same node
		healthyNodes := ring.GetNHealthyNodesForID(target["list_id"], replicationFactor) // TODO: This does not have hinted handoff into consideration

		// Send read to nodes TODO: send for a random combination of nodes to make a quorum

		var numberOfNodesToRead int = len(healthyNodes)/2 + 1

		valuesRead := make([]string, 0) // TODO: Replace with CRDT

		readChan := make(chan string)
		var readChanExpected int = 0

		for i := 0; i < numberOfNodesToRead; i++ {
			// if healthyNodes[i].Address == serverHostname && healthyNodes[i].Port == serverPort {
			// 	value := string(database.getValueRaw(target["list_id"]))
			// 	if value != "" {
			// 		valuesRead = append(valuesRead, value)
			// 	}
			// 	continue
			// }

			// node := healthyNodes[i]

			// jsonData, err := json.Marshal(map[string]string{"list_id": target["list_id"]})
			// if err != nil {
			// 	log.Printf("Error happened in JSON marshal. Err: %s", err)
			// 	readChan <- ""
			// }

			// go sendReadAndWait(node.Address, node.Port, jsonData, readChan)
			// readChanExpected += 1
		}

		for i := 0; i < readChanExpected; i++ {
			readResult := <-readChan
			if readResult != "" {
				valuesRead = append(valuesRead, readResult)
			}
		}

		// Check results obtained, merge and send back
		fmt.Println(valuesRead)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["list"] = valuesRead[0]
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		w.Write(jsonResp)

	/**
	 * This writes the data received into a key on the database
	 */
	case http.MethodPut:
		target := make(map[string]string) // TODO: Replace string with the CRDT

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}

		fmt.Println("COORDINATOR write", target)

		// The coordenator, upon receiving a write, writes locally and performs a quorum
		// however, this coordenator may not be a holder of this information, in this case
		// it only performs the quorum
		nodes := ring.GetNodes()
		healthyNodes := ring.GetNHealthyNodesForID(target["list_id"], replicationFactor) // TODO: This does not have hinted handoff into consideration

		fmt.Println("ReplicationFactor", ring.ReplicationFactor)
		// Send write to nodes TODO: send for a random combination of nodes to make a quorum
		for i := 0; i < ring.ReplicationFactor; i++ {
			virtualNodeID := healthyNodes[i]

			physicalNode := nodes[ring.ParseVirtualNodeID(virtualNodeID)[0]]

			if physicalNode.Address == serverHostname && physicalNode.Port == serverPort {
				database.writeToKey(target["list_id"], []byte(target["list"]))
				continue
			}

			payload := map[string]string{
				"list_id": target["list_id"],
				"content": target["content"],
				"node":    virtualNodeID,
			}

			jsonData, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Error happened in JSON marshal. Err: %s", err)
				continue
			}

			// FIXME: Should I wait for the response? Maybe get another node if this one fails
			protocol.SendRequestWithData(http.MethodPut, physicalNode.Address, physicalNode.Port, "/operation", jsonData)

			fmt.Println("send", payload)
		}

	}
}

func handleOperation(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	// The read operation
	case http.MethodPost:
		target := make(map[string]string)

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}

		// Get the information available on this machine
		fmt.Println("read operation", target, string(database.getValueRaw(target["list_id"])))

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["list"] = string(database.getValueRaw(target["list_id"]))
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		w.Write(jsonResp)
		return
	// The write operation
	case http.MethodPut:
		target := make(map[string]string) // TODO: Replace string with the CRDT

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}

		// Check if the data is supposed to be written in this machine or if it must be hinted handoff
		// TODO: Hinted Handoff
		fmt.Println("write operation", target)

		// Write the information received in this machine
		database.writeToKey(target["list_id"], []byte(target["list"]))
	}
}

func sendReadAndWait(address string, port string, jsonData []byte, readChan chan string) {
	response, err := protocol.SendRequestWithData(http.MethodPost, address, port, "/operation", jsonData)
	if err != nil {
		fmt.Println("Error when receiving response from read request")
		readChan <- ""
		return
	}

	target := make(map[string]string)

	err = json.NewDecoder(response.Body).Decode(&target)
	if err != nil {
		fmt.Println("Error when decoding response from read request")
		readChan <- ""
		return
	}

	readChan <- target["list"]
}
