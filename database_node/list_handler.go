package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"sdle.com/mod/crdt_go"
	"sdle.com/mod/hash_ring"
	"sdle.com/mod/protocol"
	"sdle.com/mod/utils"
)

type readChanStruct struct {
	code    int
	content *crdt_go.ShoppingList
	address string
	port    string
}

// We need this to deal with map list_id-> dotContext
type readChanStructForDotContext struct {
    code    int
    content map[string]string 
    address string
    port   
}

func handleCoordenator(w http.ResponseWriter, r *http.Request) {
	log.Println("Received /list", r.Method, "request")

	switch r.Method {

	case http.MethodPost:
		{
			target := make(map[string]string)

			decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

			if !decoded {
				return
			}

			var listId string = target["list_id"]

			// The coordenator, upon receiving a read, reads locally and performs a read quorum
			// however, this coordenator may not be a holder of this information, in this case
			// it only performs the read quorum
			healthyNodes := ring.GetHealthyNodesForID(listId)

			var healthyNodesStack utils.Stack[*hash_ring.NodeInfo]

			// Scrambles N first healthy replicas so a quorum can be performed for this key
			rand.Shuffle(min(len(healthyNodes), replicationFactor), func(i, j int) { healthyNodes[i], healthyNodes[j] = healthyNodes[j], healthyNodes[i] })

			for i := 0; i < len(healthyNodes); i++ {
				healthyNodesStack.Push(healthyNodes[i])
			}

			// Information about the success of the writes
			readChan := make(chan readChanStruct)

			var waitForRead int = 0

			// Send write to nodes
			quorumNodesNumber := min(ring.ReplicationFactor/2+1, len(healthyNodes))

			for i := 0; i < quorumNodesNumber; i++ {
				// If there aren't enough healthy nodes
				if healthyNodesStack.Size() == 0 {
					break
				}

				physicalNode := healthyNodesStack.Pop()

				payload := map[string]string{
					"list_id": listId,
				}

				go sendReadAndWait(physicalNode.Address, physicalNode.Port, payload, readChan)
				waitForRead += 1
			}

			readsContent := make([]*crdt_go.ShoppingList, 0)
			nodesRead := make([]struct {
				address string
				port    string
			}, 0)

			// TODO: TIMEOUT
			for {
				if waitForRead < 1 {
					break
				}
				result := <-readChan

				if result.code < 3 {
					if result.code == 1 {
						readsContent = append(readsContent, result.content)
					}
					nodesRead = append(nodesRead, struct {
						address string
						port    string
					}{result.address, result.port})
					waitForRead--
				} else {
					// if still has replicas
					if healthyNodesStack.Size() > 0 {
						physicalNode := healthyNodesStack.Pop()

						payload := map[string]string{
							"list_id": listId,
						}

						go sendReadAndWait(physicalNode.Address, physicalNode.Port, payload, readChan)
					} else {
						// Cannot write anymore so we do not wait
						waitForRead--
					}
				}
			}

			if len(readsContent) > 0 {
				// Merge every read
				var finalCRDT *crdt_go.ShoppingList = readsContent[0]

				for i := 1; i < len(readsContent); i++ {
					finalCRDT.Merge(readsContent[i])
				}

				// After merging
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				jsonResp, err := json.Marshal(protocol.ShoppingListOperation{ListId: listId, Content: finalCRDT})
				if err != nil {
					log.Fatalf("Error happened in JSON marshal. Err: %s", err)
				}

				w.Write(jsonResp)

				// After writing response to the user, write the final CRDT in the database
				for i := 0; i < len(nodesRead); i++ {
					go sendWrite(nodesRead[i].address, nodesRead[i].port, protocol.ShoppingListOperation{
						ListId:  listId,
						Content: finalCRDT,
					})
				}

				return
			} else if len(nodesRead) != 0 {
				w.WriteHeader(http.StatusNotFound)
				// Read but found no values
				return
			} else {
				// FAILURE: No reads were actually made
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}
	/**
	 * This writes the data received into a key on the database
	 */
	case http.MethodPut:
		{
			var target protocol.ShoppingListOperation

			decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

			if !decoded {
				return
			}

			// The coordenator, upon receiving a write, writes locally and performs a quorum
			// however, this coordenator may not be a holder of this information, in this case
			// it only performs the quorum
			healthyNodes := ring.GetHealthyNodesForID(target.ListId)

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

				go sendWriteAndWait(physicalNode.Address, physicalNode.Port, target, writeChan)
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

						go sendWriteAndWait(physicalNode.Address, physicalNode.Port, target, writeChan)
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
}

func handleOperation(w http.ResponseWriter, r *http.Request) {
	log.Println("Received /operation", r.Method, "request")
	switch r.Method {
	// The read operation
	case http.MethodPost:
		target := make(map[string]string)

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}

		// Get the information available on this machine
		valueRead, readSuccess := database.getShoppingList(target["list_id"])

		if !readSuccess {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		jsonResp, err := json.Marshal(protocol.ShoppingListOperation{ListId: target["list_id"], Content: valueRead})
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}

		w.Write(jsonResp)
		return
	// The write operation
	case http.MethodPut:
		var target protocol.ShoppingListOperation

		crdtErr, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !crdtErr {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to unmarshall CRDT"))
			return
		}

		// Write the information received in this machine
		database.updateOrSetShoppingList(target.ListId, target.Content)
	}
}

/**
 * Returns 1, 2 or 3
 * 1 - List was found and retrieved
 * 2 - No list was found
 * 3 - No response or the response is invalid
 */
func sendReadAndWait(address string, port string, payload map[string]string, readChan chan readChanStruct) {
	if address == serverHostname && port == serverPort {
		value, got_list := database.getShoppingList(payload["list_id"])

		if !got_list {
			readChan <- readChanStruct{2, nil, address, port}
			return
		}

		readChan <- readChanStruct{1, value, address, port}
		return
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s \n", err)
		readChan <- readChanStruct{3, nil, address, port}
		return
	}

	response, err := protocol.SendRequestWithData(http.MethodPost, address, port, "/operation", jsonData)
	if err != nil {
		readChan <- readChanStruct{3, nil, address, port}
		return
	}

	// Successful if read succeeds
	if response.StatusCode == http.StatusOK {
		var target protocol.ShoppingListOperation
		err = json.NewDecoder(response.Body).Decode(&target)
		if err != nil {
			fmt.Println("Error when receiving response from read request", err)
			readChan <- readChanStruct{3, nil, address, port}
			return
		}

		readChan <- readChanStruct{1, target.Content, address, port}
	} else if response.StatusCode == http.StatusNotFound {
		readChan <- readChanStruct{2, nil, address, port}
	} else {
		readChan <- readChanStruct{3, nil, address, port}
	}
}

/**
 * Returns 1, 2 or 3
 * 1 - list_id_dot_contents was found and retrieved
 * 2 - No list_id_dot_contents was found
 * 3 - No response or the response is invalid
 */
 func sendReadAndWaitDotContext(address string, port string, readChan chan readChanStructForDotContext) {
	if address == serverHostname && port == serverPort {
        listsIdDotContents, err := database.GetAllListsIdDotContents()

        if err != nil {
            
            readChan <- readChanStructForDotContext{3, nil, address, port}
            return
        }
		// No response or the response is invalid
        if len(listsIdDotContents) == 0 {
            // No list_id_dot_contents found
            readChan <- readChanStructForDotContext{2, nil, address, port}
            return
        }

        // Successfully found and retrieved lists_id_dot_contents
        readChan <- readChanStructForDotContext{1, listsIdDotContents, address, port}
        return
    }
	else{
		//AntiEntropy The adress and or port used need to be always from the hostNode when dealling with DotContext !!
		readChan <- readChanStructForDotContext{3, nil, address, port}
            return
	}
}

// Returns true if successful, false if not
func sendWriteAndWait(address string, port string, payload protocol.ShoppingListOperation, writeChan chan bool) {
	if address == serverHostname && port == serverPort {
		database.updateOrSetShoppingList(payload.ListId, payload.Content)

		writeChan <- true
		return
	}

	response, err := sendWrite(address, port, payload)
	if err != nil {
		writeChan <- false
		return
	}

	// Successful if write suceeds
	if response.StatusCode == http.StatusOK {
		writeChan <- true
	} else {
		writeChan <- false
	}
}

func sendWrite(address string, port string, payload protocol.ShoppingListOperation) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("error happened in JSON marshal: %s \n", err)
		return nil, fmt.Errorf("error happened in JSON marshal: %s", err)
	}

	return protocol.SendRequestWithData(http.MethodPut, address, port, "/operation", jsonData)
}
