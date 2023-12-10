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

		if ring.WasUpdated() {
			checkForHintedHandoff()
		}
	}
}

func gossipWith(node *hash_ring.NodeInfo) {
	if !node.GossipLock.TryLock() {
		return
	}

	gossipMaterial := ring.NodesGossip()

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
