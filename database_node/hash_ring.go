package main

import (
	"fmt"
	"sync"
)

type NodeStatus int64

const (
	NODE_OK           NodeStatus = 0 // When a Node is responsive
	NODE_UNRESPONSIVE NodeStatus = 1 // When a Node is unresponsive
	NODE_UNKNOWN      NodeStatus = 2 // When a Node is recently added to the ring, and has never been communicated before
)

type nodeInfo struct {
	id      string
	address string
	port    string
	status  NodeStatus
}

type HashRing struct {
	nodes map[string]*nodeInfo
	lock  sync.Mutex
}

/**
 * Initialize the Hash Ring
 */
func (ring *HashRing) initialize() {
	ring.nodes = make(map[string]*nodeInfo)
}

/**
 * Adds a node to the hash ring
 */
func (ring *HashRing) addNode(address string, port string) bool {
	// A node's id is made of a string of the address and the port
	var id string = fmt.Sprintf("%s:%s", address, port)

	ring.lock.Lock()

	// Checking if the node is already in the ring
	if ring.nodes[id] != nil {
		ring.lock.Unlock()
		return false
	}

	// Create the node's information struct
	var node nodeInfo = nodeInfo{
		id:      id,
		address: address,
		port:    port,
		status:  NODE_UNKNOWN,
	}

	// Add the nodeInfo to the ring
	ring.nodes[id] = &node

	ring.lock.Unlock()

	return true
}

func (ring *HashRing) getNodes() map[string]*nodeInfo {
	return ring.nodes
}
