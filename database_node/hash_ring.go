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
	id          string
	address     string
	port        string
	status      NodeStatus
	deadCounter int64
	gossipLock  sync.Mutex
}

/**
 * Creates the node information object
 */
func newNodeInfo(address string, port string, status NodeStatus) *nodeInfo {
	return &nodeInfo{
		id:          fmt.Sprintf("%s:%s", address, port),
		address:     address,
		port:        port,
		status:      status,
		deadCounter: 0,
	}
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
	if address == "" || port == "" {
		return false
	}

	// A node's id is made of a string of the address and the port
	var id string = fmt.Sprintf("%s:%s", address, port)

	ring.lock.Lock()

	// Checking if the node is already in the ring
	if ring.nodes[id] != nil {
		ring.lock.Unlock()
		return false
	}

	// Add the nodeInfo to the ring
	ring.nodes[id] = newNodeInfo(address, port, NODE_UNKNOWN)

	ring.lock.Unlock()

	return true
}

func (ring *HashRing) getNodes() map[string]*nodeInfo {
	return ring.nodes
}

func (ring *HashRing) nodesGossip() map[string][]map[string]string {
	ring.lock.Lock()
	nodesOnTheRing := ring.getNodes()

	nodesData := make(map[string][]map[string]string)
	nodesData["nodes"] = make([]map[string]string, len(nodesOnTheRing))

	for _, value := range ring.getNodes() {
		nodesData["nodes"] = append(nodesData["nodes"], map[string]string{"address": value.address, "port": value.port, "status": fmt.Sprintf("%d", value.status)})
	}

	ring.lock.Unlock()
	return nodesData
}

func (ring *HashRing) checkForNewNodes(nodes []map[string]string) {
	ring.lock.Lock()

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		if node["address"] == "" || node["port"] == "" {
			continue
		}

		if node["address"] == serverHostname && node["port"] == serverPort {
			continue
		}

		nodeId := fmt.Sprintf("%s:%s", node["address"], node["port"])

		if ring.nodes[nodeId] == nil {
			fmt.Printf("FOUND AN UNKNOWN NODE %s:%s \n", node["address"], node["port"])

			ring.nodes[nodeId] = newNodeInfo(node["address"], node["port"], NODE_UNRESPONSIVE)
		}

		// FIXME: Do I do something with the status?
	}

	ring.lock.Unlock()
}
