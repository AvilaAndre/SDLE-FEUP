package main

import (
	"encoding/json"
	"log"
	"net/http"
	"fmt"

	"sdle.com/mod/crdt_go"
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
func handleGossipPushPullAntiEntropyRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	/**
	 * Upon receiving this message, the node needs to answer the incoming gossip anti-entropy pull with 
	 with the list_id_dot_contents, comparing whith the list_ids they have in common with the sender but have different hash for the dot_context for those list_ids on reiver
	 And then the receiver send the shoppingLists with different hash(dot_context) comparing to sender
	 
	 */
	case http.MethodPost:
	{
		var incomingListIdDotContents readChanStructForDotContext
		
		decoded, incomingListIdDotContents := protocol.DecodeRequestBody(w, r.Body, incomingListIdDotContents)
		if !decoded {
			
			return
		}
		//Print the incomingListIdDotContents that comes in json format
		fmt.Printf("incomingListIdDotContents received for anti-entropy push-pull mechanism: %s", fmt.Sprintf("%v", incomingListIdDotContents))

		
		//TODO: check if here we can use/have access serverPort and serverHostname
		// Get the local node's list_id_dot_contents
		localListIdDotContentsChan := make(chan readChanStructForDotContext)

    	// Call the function with the channel
    	go sendReadAndWaitDotContext(serverHostname, serverPort, localListIdDotContentsChan)
		localListIdDotContents := <-localListIdDotContentsChan
		if localListIdDotContents.Code > 1 {
			//TODO: check if this is the best approach !
			//write message to log
			log.Printf("Error getting localListIdDotContents: %s",fmt.Sprintf("%v",localListIdDotContents))
			
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Compare hashes of dotContest between sender of pull and receiver and identify differing lists
		differingLists := make(map[string]*crdt_go.ShoppingList)
		for listId, incomingHash := range incomingListIdDotContents.Content {
			localHash, exists := localListIdDotContents.Content[listId]
			if exists && localHash != incomingHash {
				//TODO: check if here we can use/have access to serverPort and serverHostname
				payload := map[string]string{
					"list_id": listId,
				}

				shopping_list_chan := make(chan readChanStruct)
				// Here we get the local Shopping_list with listId
				go sendReadAndWait(serverHostname, serverPort, payload, shopping_list_chan)
				shopping_list := <-shopping_list_chan
				if shopping_list.code < 2 {
					differingLists[listId] = shopping_list.content
				}else{
					//TODO: check if this is the best approach !
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}else if !exists {
				//Print the shopping_list that comes in json format
				fmt.Printf("shopping_id %s received for anti-entropy push-pull mechanism, that node doenst have and will check if this list fits on any of his virtual nodes : ",listId)
				//TODO: check if this is the best approach !
				//TODO: check if the node handling Post request , have vnodes corresponding to the list_id
				//TODO: for now: continue
				continue
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
			var incoming_merged_lists map[string]*crdt_go.ShoppingList
			success, incoming_lists := protocol.DecodeRequestBody(w, r.Body, incoming_merged_lists)
			if !success {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Error decoding incoming merged lists"))
				return
			}

			incoming_merged_lists = incoming_lists
			all_success := true

			for list_id, inc_merged_list := range incoming_merged_lists {
				if !processMergedList(list_id, inc_merged_list) {
					all_success = false
				}
			}

			if all_success {
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

func processMergedList(list_id string, merged_list *crdt_go.ShoppingList) bool {
	readChan := make(chan readChanStruct)
	payload := map[string]string{"list_id": list_id}
	sendReadAndWait(serverHostname, serverPort, payload, readChan)
	result := <-readChan

	if result.code == 1 {
		local_list := result.content
		local_list.Merge(merged_list)
		return storeMergedList(list_id, local_list)
	} else if result.code == 2 {
		// If the list doesn't exist locally, just add the new list
		return storeMergedList(list_id, merged_list)
	}

	return false
}

func storeMergedList(list_id string, merged_list *crdt_go.ShoppingList) bool {
	mergedListPayload := protocol.ShoppingListOperation{
		ListId:  list_id,
		Content: merged_list,
	}
	
	writeChan := make(chan bool)
	go sendWriteAndWait(serverHostname, serverPort, mergedListPayload, writeChan)

	writeChanResult := <-writeChan
	return writeChanResult
}