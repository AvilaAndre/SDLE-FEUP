package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sdle.com/mod/utils"
	"sdle.com/mod/protocol"
	"sdle.com/mod/database_node"
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
func handleGossipPullAntiEntropyRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/**
	 * Upon receiving this message, the node needs to answer the incoming gossip anti-entropy pull with 
	 with the list_id_dot_contents, comparing whith the list_ids they have in common with the sender but have different hash for the dot_context for those list_ids on reiver
	 And then the receiver send the shoppingLists with different hash(dot_context) comparing to sender
	 
	 */
	case http.MethodPost:
	{
		
		incomingListIdDotContents := make(chan readChanStructForDotContext)// to receive list_id_dot_contents from sender pull request node

		decoded, incomingListIdDotContents := protocol.DecodeRequestBody(w, r.Body, incomingListIdDotContents)
		if !decoded {
			
			return
		}
	
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("This can try to answer the anti-entropy pull request from sender node"))
		
		//TODO: check if here we can use/have access serverPort and serverHostname
		// Get the local node's list_id_dot_contents
		localListIdDotContents := make(chan readChanStructForDotContext)// to receive Host node local list_id_dot_contents from database

		list_handler.sendReadAndWaitDotContext(serverHostname, serverPort, localListIdDotContents )
		if localListIdDotContents.code > 1 {
			//TODO: check if this is the best approach !
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Compare hashes of dotContest between sender of pull and receiver and identify differing lists
		differingLists := make(map[string]*crdt_go.ShoppingList)
		for listId, incomingHash := range incomingListIdDotContents.content {
			localHash, exists := localListIdDotContents.content[listId]
			if exists && localHash != incomingHash {
				//TODO: check if here we can use/have access to serverPort and serverHostname
				payload := map[string]string{
					"list_id": listId,
				}

				shoppingList := make(chan readChanStruct)
				// Here we get the local ShoppingList with listId
				list_handler.sendReadAndWait(serverHostname, serverPort, payload, shoppingList)
				if shoppingList.code < 2 {
					differingLists[listId] = shoppingList.content
				}else{
					//TODO: check if this is the best approach !
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		// send the differingLists
		differingListsMarshaled, err := json.Marshal(differingLists)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error marshaling differing lists"))
			return
		}
		//Push moment to the sender node in the antiEntropy mechanism
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(differingListsMarshaled)
	}

	//When sender node do push -> Sends the new merged lists to the receiver node
	case http.MethodPut:
		{
			var incomingMergedLists map[string]*crdt_go.ShoppingList
			success, incomingLists := protocol.DecodeRequestBody(w, r.Body, incomingMergedLists)
			if !success {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error decoding incoming merged lists"))
				return
			}

			incomingMergedLists = incomingLists
			allSuccess := true

			for listId, incMergedList := range incomingMergedLists {
				if !processMergedList(listId, incMergedList) {
					allSuccess = false
				}
			}

			if allSuccess {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("incoming Merged lists processed successfully"))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error processing some merged lists"))
			}
		}

	default:
		{
			protocol.WrongRequestType(w)
		}
	}
}

func processMergedList(listId string, mergedList *crdt_go.ShoppingList) bool {
	readChan := make(chan readChanStruct)
	payload := map[string]string{"list_id": listId}
	sendReadAndWait(serverHostname, serverPort, payload, readChan)
	result := <-readChan

	if result.code == 1 {
		localList := result.content
		localList.Merge(mergedList)
		return storeMergedList(listId, localList)
	} else if result.code == 2 {
		// If the list doesn't exist locally, just add the new list
		return storeMergedList(listId, mergedList)
	}

	return false
}

func storeMergedList(listId string, mergedList *crdt_go.ShoppingList) bool {
	mergedListPayload := protocol.ShoppingListOperation{
		ListId:  listId,
		Content: mergedList,
	}
	writeChan := make(chan bool)
	sendWriteAndWait(serverHostname, serverPort, mergedListPayload, writeChan)
	return <-writeChan
}