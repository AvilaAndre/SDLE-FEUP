package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"sdle.com/mod/protocol"
	"sdle.com/mod/crdt_go"
	hash_ring "sdle.com/mod/hash_ring"
	
)

func gossip() {
	for {
		for _, value := range ring.GetNodes() {
			go gossipWith(value)
		}

		time.Sleep(1 * time.Second) // FIXME: This should not be like this :p
	}
}


//TODO: Implement a more sophisticated rule here based on your requirements.
// antiEntropyInterval determines the sleep duration between anti-entropy gossip rounds.
func antiEntropyInterval(n float32) time.Duration {
	// Example: Return a fixed interval of 5 seconds
	return time.Duration(n * float32(time.Second)) // TODO: The value for n should have a concrete rule
}

// Propagate antiEntropy mechanism from sender node to a numb_choosen_nodes that a node knows
func gossipAntiEntropy(numb_choosen_nodes int32) {
	//TODO: test the code bellow
	
	for {
		var server_host_node_id string = fmt.Sprintf("%s:%s", serverHostname, serverPort)
		var server_host_vnodes_hash_keys = make([]string, 0)

		
		// Get all vnodes hash keys from the ring of the current node
		for index := 0; index < len(ring.GetNodes()[server_host_node_id].vnodes); index++ { //We directly use ring on getNodes because we dont want to clone ring that have a Mutex parameter
			var hash_id_vnode = hashId(ring.GetNodes()[server_host_node_id].vnodes[index])
			server_host_vnodes_hash_keys = append(server_host_vnodes_hash_keys, hash_id_vnode)
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
		// Check if there are enough nodes to choose from
		if int32(len(allNodes)) < numb_choosen_nodes {
			log.Printf("Not enough nodes to perform anti-entropy gossip")
			time.Sleep(antiEntropyInterval(10.0)) // Wait longer if not enough nodes
			continue
		}
		src := rand.NewSource(time.Now().UnixNano())
		r := rand.New(src)
		
		// Shuffle allNodes to randomize the selection and select a subset of nodes for anti-entropy gossip
		r.Shuffle(min(len(allNodes),replicationFactor), func(i, j int) {
			allNodes[i], allNodes[j] = allNodes[j], allNodes[i]
		})

		
		selectedNodes := allNodes[:numb_choosen_nodes]

		
		for _, node := range selectedNodes {
			go gossipAntiEntropyWith(node,3) //TODO: check if 3 is a good number of tries in this type of "gossip"
		}

		// Sleep for a defined interval before the next gossip round
		time.Sleep(antiEntropyInterval(2.0)) // TODO: Check this

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
	return
}


//Push-pull gossip dot context ( awset ) anti-entropy mechanism: Pull side
func gossipAntiEntropyWith(node *hash_ring.NodeInfo,  numb_tries int64) {
	if !node.GossipLock.TryLock() { //TODO: check if this is important to check ?
		return
	}
	

	// Here we need to read all list_ids and dot_context for every list and save on a Map
	read_chan_with_dot_context := make(chan readChanStructForDotContext)// to receive ShoppingLists from database

	// Get all the list_id_dot_contents Map from the sender/Host node 
	sendReadAndWaitDotContext(serverHostname, serverPort, read_chan_with_dot_context)
	

	// send all the list_id_dot_contents to node and wait for response
	//TODO: 
	response_from_pull, err := protocol.SendRequestWithData(http.MethodPost, node.Address, node.Port, "/gossip/antiEntropy/request/pull", read_chan_with_dot_context)
	
	if err != nil {
        handleCommunicationError(node, numb_tries)
		node.GossipLock.Unlock()
        return
    }

	// Await for response to get the node's status and shoppingLists to merge that differ with Host node on the dot context
	if response_from_pull.StatusCode == http.StatusOK  {
		node.DeadCounter = 0
		if node.Status != hash_ring.NODE_OK {
			node.Status = hash_ring.NODE_OK
			log.Printf("Node %s is OK\n", node.Id)
			log.Printf("Received anti-entropy data from node %s\n")
		}

		handleSuccessfulPullPushResponse(node, response_from_pull)
		
    } else {
		handleCommunicationError(node, numb_tries)
		fmt.Println("Non-OK status code received:", response_from_pull.StatusCode)
	}
    

	node.GossipLock.Unlock()
	return	
}



func handleCommunicationError(node *hash_ring.NodeInfo, numb_tries int64) {
    if node.DeadCounter < numb_tries {
        node.DeadCounter++
    } else if node.DeadCounter == numb_tries {
        node.Status = hash_ring.NODE_UNRESPONSIVE
        log.Printf("%s set to UNRESPONSIVE when trying anti-entropy mechanism\n", node.Id)
        node.DeadCounter++
    }
}

func handleSuccessfulPullPushResponse(node *hash_ring.NodeInfo, response *http.Response) {


differing_lists := make(map[string]*crdt_go.ShoppingList)
success, decoded_differingLists := protocol.DecodeHTTPResponse(nil, response, differing_lists)
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
    response_from_push, err := protocol.SendRequestWithData(http.MethodPut, node.Address, node.Port, "/gossip/antiEntropy/response/push", marshaled_merged_lists)
    if err != nil {
        fmt.Println("Error sending merged lists to receiver node:", err)
    } else if response_from_push.StatusCode != http.StatusOK {
        fmt.Println("Non-OK status code received from receiver node during push:", response_from_push.StatusCode)
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
			var local_list *crdt_go.ShoppingList = result.content
            local_list.Merge(common_list) //local_list merge with common_list
            merged_lists[list_id] = local_list

            // Store the merged list back into the local node's database
          	err :=  storeMergedList(list_id, local_list)
			if !err {
				fmt.Println("Error storing merged list with ID:", list_id)

			}
    	}
	}
	return merged_lists
}
// func storeMergedList(list_id string, merged_list *crdt_go.ShoppingList) {
//     merged_list_payload := protocol.ShoppingListOperation{
//         ListId:  list_id,
//         Content: merged_list,
//     }
//     write_chan := make(chan bool)
//     sendWriteAndWait(serverHostname, serverPort, merged_list_payload,write_chan)
//     if success := <-write_chan; !success {
//         fmt.Println("Error storing merged list with ID:", list_id)
//     }
// }