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

		if ring.WasUpdated() {
			checkForHintedHandoff()
		}
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
		for index := 0; index < len(ring.GetNodes()[server_host_node_id].Vnodes); index++ { //We directly use ring on getNodes because we dont want to clone ring that have a Mutex parameter
			var hash_id_vnode = hash_ring.HashId(ring.GetNodes()[server_host_node_id].Vnodes[index])
			server_host_vnodes_hash_keys = append(server_host_vnodes_hash_keys, hash_id_vnode)
		}

		var allReplicationNodes [][]*hash_ring.NodeInfo

		
		for _, vnodeHashKey := range server_host_vnodes_hash_keys {
			replicationNodes := ring.GetHealthyNodesForID(vnodeHashKey)
			if len(replicationNodes) > ring.ReplicationFactor {
				replicationNodes = replicationNodes[:ring.ReplicationFactor]
			}
			allReplicationNodes = append(allReplicationNodes, replicationNodes)
		}
		


        
       
	
		// Check if there are enough nodes to choose from
		if int32(len(allReplicationNodes)) < numb_choosen_nodes {
			log.Printf("Not enough nodes to perform anti-entropy gossip")
			time.Sleep(antiEntropyInterval(10.0)) // Wait longer if not enough nodes
			continue
		}
		src := rand.NewSource(time.Now().UnixNano())
		r := rand.New(src)
		
		// Shuffle allReplicationNodes to randomize the selection and select a subset of nodes for anti-entropy gossip
		r.Shuffle(min(len(allReplicationNodes),replicationFactor), func(i, j int) {
			allReplicationNodes[i], allReplicationNodes[j] = allReplicationNodes[j], allReplicationNodes[i]
		})

		
		selectedNodes := allReplicationNodes[:numb_choosen_nodes]

		
		for _, replicationNodeGroup := range selectedNodes {
			for _, node := range replicationNodeGroup {
				go gossipAntiEntropyWith(node, 3) 
			}
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
			ring.NodeStatusChanged()

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
			ring.NodeStatusChanged()
		}
	}

	node.GossipLock.Unlock()
	
}


//Push-pull gossip dot context ( awset ) anti-entropy mechanism: Pull side
func gossipAntiEntropyWith(node *hash_ring.NodeInfo,  numb_tries int64) {
	if !node.GossipLock.TryLock() { //TODO: check if this is important to check ?
		return
	}
	

	// Here we need to read all list_ids and dot_context for every list and save on a Map
	read_chan_with_dot_context_chan := make(chan readChanStructForDotContext)// to receive ShoppingLists from database

	// Get all the list_id_dot_contents Map from the sender/Host node 
	go sendReadAndWaitDotContext(serverHostname, serverPort, read_chan_with_dot_context_chan)
	read_chan_with_dot_context := <-read_chan_with_dot_context_chan
	
	jsonDotContext, err := json.Marshal(read_chan_with_dot_context)

	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)

		node.GossipLock.Unlock()
		return
	}

	// send all the list_id_dot_contents to node and wait for response
	//TODO: 
	response_from_pull, err := protocol.SendRequestWithData(http.MethodPost, node.Address, node.Port, "/gossip/antiEntropy/request/pull", jsonDotContext)
	
	if err != nil {
        handleCommunicationError(node)
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
		handleCommunicationError(node)
		fmt.Println("Non-OK status code received:", response_from_pull.StatusCode)
	}
    

	node.GossipLock.Unlock()
	
}



func handleCommunicationError(node *hash_ring.NodeInfo) {
    fmt.Println("SendRequestData with dotContext failed to node %s: %s", node.Port)
}

func handleSuccessfulPullPushResponse(node *hash_ring.NodeInfo, response *http.Response) {


	differing_lists := make(map[string]*crdt_go.ShoppingList)
	success, decoded_differingLists := protocol.DecodeHTTPResponse(nil, response, differing_lists)//TODO: check this if decoded properly
	if !success {
		fmt.Println("Error decoding response from anti-entropy pull request")
		return
	}

	differing_lists = decoded_differingLists
	merged_lists,err := processDifferingLists(differing_lists)

	if err != nil {
		// Handle the error
		fmt.Printf("Error processing differing lists: %s\n", err)
		return
	}

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

func processDifferingLists(differing_lists map[string]*crdt_go.ShoppingList) (map[string]*crdt_go.ShoppingList, error) {
    merged_lists := make(map[string]*crdt_go.ShoppingList)
	var err error

    for list_id, common_list := range differing_lists {
        readChan := make(chan readChanStruct)
        payload := map[string]string{"list_id": list_id}
        go sendReadAndWait(serverHostname, serverPort, payload, readChan)
        result_read := <-readChan
        
        switch result_read.code {
			case 1: // List was found and retrieved
				local_list := result_read.content
				local_list.Merge(common_list) // Merge local_list with common_list
				merged_lists[list_id] = local_list
	
				// Store the merged list back into the local node's database
				if !storeMergedList(list_id, local_list) {
					fmt.Printf("Error storing merged list with ID: %s\n", list_id)
				}
	
			case 2: // No list was found but is supposed to exist
				fmt.Printf("List with ID %s not found locally\n", list_id)
				err = fmt.Errorf("list with ID %s not found locally", list_id)
	
			case 3: // No response or the response is invalid
				fmt.Printf("Invalid response or error occurred when fetching list with ID %s\n", list_id)
				err = fmt.Errorf("invalid response or error occurred when fetching list with ID %s", list_id)
			}
			
		
		}
	
		return merged_lists, err
}
