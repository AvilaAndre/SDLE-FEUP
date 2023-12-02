package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"sdle.com/mod/hash_ring"
	"sdle.com/mod/protocol"
	"sdle.com/mod/utils"
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
		healthyNodes := ring.GetHealthyNodesForID(target["list_id"]) // TODO: This does not have hinted handoff into consideration

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

		// The coordenator, upon receiving a write, writes locally and performs a quorum
		// however, this coordenator may not be a holder of this information, in this case
		// it only performs the quorum
		healthyNodes := ring.GetHealthyNodesForID(target["list_id"])

		var healthyNodesStack utils.Stack[*hash_ring.NodeInfo]

		// Scrambles N first healthy replicas so a quorum can be performed for this key
		rand.Shuffle(min(len(healthyNodes), replicationFactor), func(i, j int) { healthyNodes[i], healthyNodes[j] = healthyNodes[j], healthyNodes[i] })

		for i := 0; i < len(healthyNodes); i++ {
			healthyNodesStack.Push(healthyNodes[i])
		}

		// Information about the success of the writes
		writeChan := make(chan bool)

		var waitForWrite int = 0
		var wroteSuccessfully int = 0

		// Send write to nodes
		quorumNodesNumber := min(ring.ReplicationFactor/2+1, len(healthyNodes))

		for i := 0; i < quorumNodesNumber; i++ {
			// If there aren't enough healthy nodes
			if healthyNodesStack.Size() == 0 {
				break
			}

			physicalNode := healthyNodesStack.Pop()

			payload := map[string]string{
				"list_id": target["list_id"],
				"content": target["content"],
			}

			go sendWriteAndWait(physicalNode.Address, physicalNode.Port, payload, writeChan)
			waitForWrite += 1
		}

		// TODO: TIMEOUT
		for {
			if waitForWrite < 1 {
				break
			}
			result := <-writeChan

			if result {
				wroteSuccessfully++
				waitForWrite--
			} else {
				// if still has replicas
				if healthyNodesStack.Size() > 0 {
					physicalNode := healthyNodesStack.Pop()

					payload := map[string]string{
						"list_id": target["list_id"],
						"content": target["content"],
					}

					go sendWriteAndWait(physicalNode.Address, physicalNode.Port, payload, writeChan)
				} else {
					// Cannot write anymore so we do not wait
					waitForWrite--
				}
			}
		}

		if wroteSuccessfully > 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		return
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

// Returns true if successful, false if not
func sendWriteAndWait(address string, port string, payload map[string]string, writeChan chan bool) {
	if address == serverHostname && port == serverPort {
		database.writeToKey(payload["list_id"], []byte(payload["content"]))

		writeChan <- true
		return
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Panicf("Error happened in JSON marshal. Err: %s \n", err)
		writeChan <- false
		return
	}

	response, err := protocol.SendRequestWithData(http.MethodPut, address, port, "/operation", jsonData)
	if err != nil {
		writeChan <- false
		return
	}

	// Successful if write suceeds
	if response.StatusCode == 200 {
		writeChan <- true
	} else {
		writeChan <- false
	}
}
