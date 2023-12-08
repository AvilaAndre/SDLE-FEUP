package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	hash_ring "sdle.com/mod/hash_ring"
	"sdle.com/mod/protocol"
)

func gossip() {
	for {
		for _, value := range ring.GetNodes() {
			go gossipWith(value)
		}

		time.Sleep(1 * time.Second) // FIXME: This should not be like this :p
	}
}


// antiEntropyInterval determines the sleep duration between anti-entropy gossip rounds.
//TODO: Implement a more sophisticated rule here based on your requirements.
func antiEntropyInterval(float32 n) time.Duration {
    // Example: Return a fixed interval of 5 seconds
    return n * time.Second // TODO: The value for n should have a concrete rule
}

// Propagate antiEntropy mechanism with a numb_choosen_nodes that a node knows
func gossipAntiEntropy(numb_choosen_nodes int) {
	//TODO: test the code bellow
	
	for {
		var server_host_node_id string = fmt.Sprintf("%s:%s", serverHostname, serverPort)
		var server_host_vnodes_ids = ring.nodes[server_host_node_id].vnodes
		var server_host_vnodes_hash_keys = make([]string, 0)

		for index := 0; index < len(server_host_vnodes_ids); index++ {
			var hash_id_vnode = ring.hashId(server_host_vnodes_ids[index])
			server_host_vnodes_hash_keys = append(server_host_vnodes_hash_key, hash_id_vnode)
		}

		var allReplicationNodes [][]*hash_ring.NodeInfo

		// Get all replication nodes from the ring of the current node: who is doing the gossipAntiEntropy
		for _, vnodeHashKey := range server_host_vnodes_hash_keys {
			replicationNodes := ring.GetHealthyNodesForID(vnodeHashKey)
			if len(replicationNodes) > ring.ReplicationFactor {
				replicationNodes = replicationNodes[:ring.ReplicationFactor]
			}
			allReplicationNodes = append(allReplicationNodes, replicationNodes)
		}
		


        
       
		var allNodes []*hash_ring.NodeInfo
		for _, replicationNodes := range allReplicationNodes {
			allNodes = append(allNodes, replicationNodes...)
		}

		// Check if there are enough nodes to choose from
		if len(allNodes) < numb_choosen_nodes {
			log.Printf("Not enough nodes to perform anti-entropy gossip")
			time.Sleep(antiEntropyInterval(10.0)) // Wait longer if not enough nodes
			continue
		}

		// Shuffle allNodes to randomize the selection and select a subset of nodes for anti-entropy gossip
		rand.Shuffle(len(allNodes), func(i, j int) {
			allNodes[i], allNodes[j] = allNodes[j], allNodes[i]
		})

		
		selectedNodes := allNodes[:numb_choosen_nodes]

		
		for _, node := range selectedNodes {
			go gossipAntiEntropyWith(node)
		}

		// Sleep for a defined interval before the next gossip round
		time.Sleep(antiEntropyInterval(2.0)) // Check this

    }
}



func gossipWith(node *hash_ring.NodeInfo) {
	if !node.GossipLock.TryLock() {
		return
	}

	gossipMaterial := ring.NodesGossip()//TODO: The ring nodes of the node who call gossip? Ask Avila

	jsonData, err := json.Marshal(gossipMaterial)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		node.GossipLock.Unlock()
		return
	}

	response, err2 := protocol.SendRequestWithData(http.MethodPost, node.Address, node.Port, "/gossip", jsonData)
	if err2 != nil {
		// If cannot gossip four consecutive times then assume the node is dead
		if node.DeadCounter < 3 {
			node.DeadCounter++
		} else if node.DeadCounter == 3 {
			node.Status = hash_ring.NODE_UNRESPONSIVE
			log.Printf("%s set to UNRESPONSIVE\n", node.Id)
			node.DeadCounter++
		}

		node.GossipLock.Unlock()
		return
	}

	// Await for response to get the node's status
	if response.StatusCode == 200 {
		node.DeadCounter = 0
		if node.Status != hash_ring.NODE_OK {
			node.Status = hash_ring.NODE_OK
			log.Printf("%s set to OK\n", node.Id)
		}
	}

	node.GossipLock.Unlock()
}


//Push-pull gossip dot context ( awset ) anti-entropy mechanism: Pull side
func gossipAntiEntropyWith(node *hash_ring.NodeInfo, uint32 numb_tries) {
	if !node.GossipLock.TryLock() { //TODO: check if this is important to check ?
		return
	}
	

	// Here we need to read all list_ids and dot_context for every list and save on a Map
	read_chan_with_dot_context := make(chan readChanStructForDotContext)// to receive ShoppingLists from database

	// Get all the list_id_dot_contents Map from the sender/Host node 
	list_handler.sendReadAndWaitDotContext(serverHostname, serverPort, read_chan_with_dot_context)
	

	// send all the list_id_dot_contents to node and wait for response
	//TODO: 
	response_from_pull, err := protocol.SendRequestWithData(http.MethodPost, node.Address, node.Port, "/gossip/antiEntropy/request/pull", read_chan_with_dot_context)
	
	if err != nil {
        handleCommunicationError(node, numb_tries)
        return
    }

	// Await for response to get the node's status and shoppingLists to merge that differ with Host node on the dot context
	if response_from_pull.StatusCode == http.StatusOK  {
		node.DeadCounter = 0
		if node.Status != hash_ring.NODE_OK {
			node.Status = hash_ring.NODE_OK
			log.Printf("Node %s is OK\n", node.Id)
			log.Printg("Received anti-entropy data from node %s\n")
		}

		handleSuccessfulPullPushResponse(node, response_from_pull)
    } 
	else {
        fmt.Println("Non-OK status code received:", response_from_pull.StatusCode)
    }

	node.GossipLock.Unlock()
		

	
}

func handleCommunicationError(node *hash_ring.NodeInfo, numb_tries uint32) {
    if node.DeadCounter < numb_tries {
        node.DeadCounter++
    } else if node.DeadCounter == numb_tries {
        node.Status = hash_ring.NODE_UNRESPONSIVE
        log.Printf("%s set to UNRESPONSIVE when trying anti-entropy mechanism\n", node.Id)
        node.DeadCounter++
    }
    node.GossipLock.Unlock()
}

func handleSuccessfulPullPushResponse(node *hash_ring.NodeInfo, response *http.Response) {
    differing_lists := make(map[string]*crdt_go.ShoppingList)
    success, decoded_differingLists := DecodeHTTPResponse(nil, response, differing_lists)
    if !success {
        fmt.Println("Error decoding response from anti-entropy pull request")
        return
    }

    differing_lists = decoded_differingLists
    merged_lists := processDifferingLists(differing_lists)

    // Send the merged new shoppingLists to the receiver node that have responded
    marshaled_merged_lists, err := json.Marshal(merged_lists)
    if err != nil {
        fmt.Println("Error marshaling merged lists:", err)
        return
    }
	//TODO: check if i need to retun on err !=nill
    // Finally we send the merged shoppingLists, requesting a push in the anti-entropy mechanism
    response_from_push, err := protocol.SendRequestWithData(http.MethodPut, node.Address, node.Port, "/gossip/antiEntropy/response/push", marshaledMergedLists)
    if err != nil {
        fmt.Println("Error sending merged lists to receiver node:", err)
    } else if response_from_push.StatusCode != http.StatusOK {
        fmt.Println("Non-OK status code received from receiver node during push:", responseFromPush.StatusCode)
    }

	if response_from_push.StatusCode == http.StatusOK{
		fmt.Println("AntiEntropy pushpull dot context mechanism totally successful!")
	} 
}

func processDifferingLists(differing_lists map[string]*crdt_go.ShoppingList) map[string]*crdt_go.ShoppingList {
    merged_lists := make(map[string]*crdt_go.ShoppingList)
    for list_id, common_list := range differing_lists {
        readChan := make(chan readChanStruct)
        payload := map[string]string{"list_id": list_id}
        sendReadAndWait(serverHostname, serverPort, payload, readChan)
        result := <-readChan
        
        if result.code == 1 {
            local_list := result.content
            merged_list := localList.Merge(common_list)
            merged_lists[list_id] = merged_list

            // Store the merged list back into the local node's database
            storeMergedList(list_id, merged_list)
        }
    }
    return merged_lists
}

func storeMergedList(list_id string, merged_list *crdt_go.ShoppingList) {
    merged_list_payload := protocol.ShoppingListOperation{
        ListId:  list_id,
        Content: merged_list,
    }
    write_chan := make(chan bool)
    sendWriteAndWait(serverHostname, serverPort, merged_list_payload,write_chan)
    if success := <-write_chan; !success {
        fmt.Println("Error storing merged list with ID:", list_id)
    }
}