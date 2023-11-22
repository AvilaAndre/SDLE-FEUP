package main

import "sync"

type node struct {
	address string
	port    string
}

type nodeManager struct {
	nodes []node
	Mu    sync.Mutex
}

func (manager *nodeManager) hasNode(newNode node) bool {
	manager.Mu.Lock()
	for i := 0; i < len(manager.nodes); i++ {
		node := manager.nodes[i]

		if node.address == newNode.address && node.port == newNode.port {
			manager.Mu.Unlock()
			return true
		}
	}
	manager.Mu.Unlock()
	return false
}

func (manager *nodeManager) addNode(newNode node) bool {
	if manager.hasNode(newNode) {
		return false
	}

	manager.Mu.Lock()

	manager.nodes = append(manager.nodes, newNode)

	manager.Mu.Unlock()

	return true
}

func (manager *nodeManager) getNodes() []node {
	manager.Mu.Lock()
	defer manager.Mu.Unlock()
	return manager.nodes
}
