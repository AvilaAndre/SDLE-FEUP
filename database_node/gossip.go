package main

import (
	
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"sdle.com/mod/crdt_go"
	hash_ring "sdle.com/mod/hash_ring"
	"sdle.com/mod/protocol"
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
	fmt.Println("Anti-entropy mechanism started!")
	for {
		var server_host_node_id string = fmt.Sprintf("%s:%s", serverHostname, serverPort)
		var server_host_vnodes_hash_keys = make([]string, 0)
		

		
		// Get all vnodes hash keys from the ring of the current node
		for index := 0; index < len(ring.GetNodes()[server_host_node_id].Vnodes); index++ { //We directly use ring on getNodes because we dont want to clone ring that have a Mutex parameter
			var vnode_id = ring.GetNodes()[server_host_node_id].Vnodes[index]
			var vnode_hash string = hash_ring.HashId(vnode_id)
			server_host_vnodes_hash_keys = append(server_host_vnodes_hash_keys, vnode_hash)
		}

		var allReplicationNodes [][]*hash_ring.NodeInfo
		// print server_host_vnodes_hash_keys
		fmt.Println("Server host vnodes hash keys on the ring:", server_host_vnodes_hash_keys)
		
		for _, vnodeHashKey := range server_host_vnodes_hash_keys {
			replicationNodes := ring.GetHealthyNodesForID(vnodeHashKey)
			if len(replicationNodes) > ring.ReplicationFactor {
				replicationNodes = replicationNodes[:ring.ReplicationFactor]
			}
			allReplicationNodes = append(allReplicationNodes, replicationNodes)
		}
		
		
		fmt.Println("All replication nodes possible to do anti-entropy:", allReplicationNodes)

        
		src := rand.NewSource(time.Now().UnixNano())
		r := rand.New(src)
		
		// Shuffle allReplicationNodes to randomize the selection and select a subset of nodes for anti-entropy gossip
		r.Shuffle(min(len(allReplicationNodes),ring.ReplicationFactor), func(i, j int) {
			allReplicationNodes[i], allReplicationNodes[j] = allReplicationNodes[j], allReplicationNodes[i]
		})

		// Flatten and deduplicate allReplicationNodes
		nodeMap := make(map[string]*hash_ring.NodeInfo)
		for _, nodes := range allReplicationNodes {
			for _, node := range nodes {
				nodeMap[node.Id] = node
			}
		}

		flattenedNodes := make([]*hash_ring.NodeInfo, 0, len(nodeMap))
		for _, node := range nodeMap {
			flattenedNodes = append(flattenedNodes, node)
		}
		r.Shuffle(len(flattenedNodes), func(i, j int) {
			flattenedNodes[i], flattenedNodes[j] = flattenedNodes[j], flattenedNodes[i]
		})
		//print flattenedNodes
		fmt.Println("Flattened nodes to do anti-entropy:", flattenedNodes)
		
		fmt.Println("Number of nodes to do anti-entropy:", numb_choosen_nodes)
		//print length of flattenedNodes
		fmt.Println("Length of flattened nodes to do anti-entropy:", len(flattenedNodes))
		// decalare selectedNodes
		
		selectedNodes := flattenedNodes
		if numb_choosen_nodes < int32(len(flattenedNodes)) {
			selectedNodes = flattenedNodes[:numb_choosen_nodes]
		}
		
	
		
		//print value of numb_choosen_nodes
		fmt.Println("Selected nodes to do anti-entropy:", selectedNodes)
		
		for _, node := range selectedNodes {
			
				//print going to do anti-entropy with node
				fmt.Println("I am node port: ", serverPort," and address: ", serverHostname)
				fmt.Println("Going to try send dot context anti-entropy with node to port:", node.Port, " and address:", node.Address)
				//Dont do anti entropy with yourself crazy node!
				if node.Address == serverHostname && node.Port == serverPort{
					fmt.Println("You are not going to do dot context anti-entropy with yourself!")
					
					continue
				}
				go gossipAntiEntropyWith(node, 2) 
			
		}

		// Sleep for a defined interval before the next gossipAntiEntropy round
		time.Sleep(antiEntropyInterval(60.0)) // TODO: Check this
		//Print one gossip round done
		fmt.Println("One dotContext antiEntropy push-pull gossip round done!!!!!")

    }
}

func HashId(vnode_id string) {
	panic("unimplemented")
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
		//print this("Node is locked and ready for dot context anti-entropy ")
		fmt.Println("Node is locked and ready for dot context anti-entropy")

		return
	}
	
	
	fmt.Println("I am the sender node and i will send dot context to receiver node: for dot context pull anti-entropy")
	// Here we need to read all list_ids and dot_context for every list and save on a Map
	read_chan_with_dot_context_chan := make(chan readChanStructForDotContext)// to receive ShoppingLists from database

	// Get all the list_id_dot_contents Map from the sender/Host node 
	go sendReadAndWaitDotContext(serverHostname, serverPort, read_chan_with_dot_context_chan)
	read_chan_with_dot_context := <-read_chan_with_dot_context_chan
	
	if read_chan_with_dot_context.Code != 1 {
		// dot context was not found
		fmt.Println("Ids Dot context doesnt exist on sender node, no anti-entropy action needed.")
		node.GossipLock.Unlock()
		return
	}


	fmt.Println("I have the dot context to send to receiver node on anti entropy:", read_chan_with_dot_context.Content)
	jsonDotContext, err := json.Marshal(read_chan_with_dot_context)
	
	if len(read_chan_with_dot_context.Content) == 0 {
        fmt.Println("The dot context map is empty, no data to send for anti-entropy!.")
        node.GossipLock.Unlock()
        return
    }
	// fmt.Println("I have the json Marshall dot context to send to receiver node:", jsonDotContext)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)

		node.GossipLock.Unlock()
		return
	}
	// send all the list_id_dot_contents to node and wait for response
	//TODO: 
	response_from_pull, err := protocol.SendRequestWithData(http.MethodPost, node.Address, node.Port, "/gossip/antiEntropy/request", jsonDotContext)
	
	if err != nil {
		
		log.Printf("Error sending request: %s", err)
		return
	}

	bodyBytes, err := ioutil.ReadAll(response_from_pull.Body)
	responseBody := string(bodyBytes)
	if responseBody == "No differing lists" {
		fmt.Println("No differing lists found on receiver node, no anti entropy action needed.")
		node.GossipLock.Unlock()
		return
	}
	fmt.Println("Response from receiver node to pull anti-entropy:", responseBody)
	if err != nil {
		
		log.Printf("Error reading response body: %s", err)
		return
	}


	fmt.Println("I have the response_from_pull.Body from receiver node on anti entropy:", responseBody)
	if err != nil {
        handleCommunicationError(node)
		node.GossipLock.Unlock()
        return
    }

	// Await for response to get the node's status and shoppingLists to merge that differ with Host node on the dot context
	if response_from_pull.StatusCode == http.StatusOK  {
		

		handleSuccessfulPullPushResponse(node, response_from_pull)
		//print sender node that anti-entropy pull-push was successful
		fmt.Println("Function: handleSuccessfulPullPushResponse -> Anti-entropy pull-push was successful!")
    } else {
		handleCommunicationError(node)
		fmt.Println("Non-OK status code received:", response_from_pull.StatusCode)
		node.GossipLock.Unlock()
	}
    
	//Print successful anti-entropy pull-push
	fmt.Println("End of one anti-entropy pull-push mechanism go routine !!!!")
	node.GossipLock.Unlock()
	
}



func handleCommunicationError(node *hash_ring.NodeInfo) {

	fmt.Printf("SendRequestData with dotContext failed to node %s\n", node.Port)
}

func handleSuccessfulPullPushResponse(node *hash_ring.NodeInfo, response *http.Response) {
	
	bodyBytes, err := ioutil.ReadAll(response.Body)
	responseBody := string(bodyBytes)
	differing_lists := make(map[string]*crdt_go.ShoppingList)
	success, decoded_differingLists := protocol.DecodeHTTPResponse(nil, response, differing_lists)//TODO: check this if decoded properly
	if responseBody == "No differing lists" {
		fmt.Println("No differing lists found on receiver node, no anti entropy action needed.")
		node.GossipLock.Unlock()
		return
	}
	
	fmt.Printf("Decoded_differingLists on receiver side for anti-entropy push-pull mechanism: %s", fmt.Sprintf("%v", decoded_differingLists))
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
	response_from_push, err := protocol.SendRequestWithData(http.MethodPut, node.Address, node.Port, "/gossip/antiEntropy/request", marshaled_merged_lists)
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
