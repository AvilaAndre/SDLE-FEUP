package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"sdle.com/mod/protocol"
)

func gossip() {
	for {
		for _, value := range ring.getNodes() {
			go gossipWith(value)
		}

		time.Sleep(1 * time.Second) // FIXME: This should not be like this :p
	}
}

func gossipWith(node *nodeInfo) {
	if !node.gossipLock.TryLock() {
		return
	}

	gossipMaterial := ring.nodesGossip()

	jsonData, err := json.Marshal(gossipMaterial)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		node.gossipLock.Unlock()
		return
	}

	response, err2 := protocol.SendRequestWithData(http.MethodPost, node.address, node.port, "/gossip", jsonData)
	if err2 != nil {
		// If cannot gossip four consecutive times then assume the node is dead
		if node.deadCounter < 3 {
			node.deadCounter++
		} else if node.deadCounter == 3 {
			node.status = NODE_UNRESPONSIVE
			log.Printf("%s set to UNRESPONSIVE\n", node.id)
			node.deadCounter++
		}

		node.gossipLock.Unlock()
		return
	}

	node.gossipLock.Unlock()

	// Await for response to get the node's status
	if response.StatusCode == 200 {
		node.deadCounter = 0
		if node.status == NODE_UNRESPONSIVE {
			node.status = NODE_OK
			log.Printf("%s set to OK\n", node.id)
		}
	}
}
