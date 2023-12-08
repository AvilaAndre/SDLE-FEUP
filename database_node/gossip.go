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
	if !node.GossipLock.TryLock() {
		return
	}
	//network addresses, ports, etc from all nodes that the current node who call gossipAntiEntropyWith knows
	// format returned bellow : nodesData["nodes"] = append(nodesData["nodes"], map[string]string{"address": value.Address, "port": value.Port, "status": fmt.Sprintf("%d", value.Status)})
	antiEntropyGossipMaterial := ring.NodesGossip()//TODO: ask if with this i know the server, port of other nodes because of the gossip backend process that is also running every few seconds !

	jsonData, err := json.Marshal(antiEntropyGossipMaterial)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		node.GossipLock.Unlock()
		return
	}
	//Send JSON of map listsIdsAwsets : list_id -> hash(awset_of_list_list_id) to a given node
	//TODO: listsIdsAwsets := ownListsIdsHashAwsets ->info on the node who do the gossipAntiEntropyWith
	//Send also jsonData with the gossipMaterial needed
	// "/gossip/antiEntropy" indicates to use handleGossipAntiEntropy method on registerRoutes of database_node
	response, err2 := protocol.SendRequestWithAntiEntropyData(http.MethodPost, node.Address, node.Port, "/gossip/antiEntropy", jsonData, listsIdsHashAwsets)
	if err2 != nil {
		// If cannot gossip with anti-entropy message 3 consecutive times then assume the node is dead
		if node.DeadCounter < numb_tries { //TODO: check this later ?
			node.DeadCounter++
		} else if node.DeadCounter == numb_tries { //The basic gossip event will deal Nodes unresponsive
			node.Status = hash_ring.NODE_UNRESPONSIVE
			log.Printf("%s set to UNRESPONSIVE when trying anti-entropy mechanism\n", node.Id)
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
			log.Printf("Node %s is OK\n", node.Id)
			log.Printg("Received anti-entropy data from node %s\n")
		}
		// TODO:Continue the anti-entropy Mechanism here or outside in gossip/gossipEntropy function ??
		
		//Sending the requested ShoppingLists and possibly updating the received those ShoppingLists
		//TODO: do the merge with the received ShoppingLists and save the list_id of incoming ShoppingLists in variable list_ids

			// decode the body response and get the shoppingLists

			// create a map: incomingShoppingLists with list_id -> CRDTOfShopping_list_id of the shoppingLists received

			// for the list_ids in the map incomingShoppingLists: get the sender node stored common ShoppingLists

			//merge the incoming_lists with the sender node common lists

			// sender node store the new merged ShoppingLists he have in common with receiver node

			/* sender sends the in common with receiver node the new merged ShoppingLists,  and waits for response.StatusCode 
			( receiver node then merge the received new merged shoppingLists from sender node) */


			// 
		//TODO: 
		// Use the newly merged ShoppingLists above and send them to the receiver node

		response2, err3 := protocol.SendRequestWithAntiEntropyData(http.MethodPost, node.Address, node.Port, "/gossip/antiEntropy/response", jsonData, listsIdsHashAwsets)
	
	}

	node.GossipLock.Unlock()
}