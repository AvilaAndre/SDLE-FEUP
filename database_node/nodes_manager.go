package main

type node struct {
	address string
	port    string
}

type nodeManager struct {
	nodes []node
}

func (manager nodeManager) hasNode(newNode node) bool {
	for i := 0; i < len(manager.nodes); i++ {
		node := manager.nodes[i]

		if node.address == newNode.address && node.port == newNode.port {
			return true
		}
	}
	return false
}

func (manager *nodeManager) addNode(newNode node) bool {
	if manager.hasNode(newNode) {
		return false
	}

	manager.nodes = append(manager.nodes, newNode)

	return true
}

func (manager nodeManager) getNodes() []node {
	return manager.nodes
}
