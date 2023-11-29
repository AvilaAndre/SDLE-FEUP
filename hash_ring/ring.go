package hash_ring

import (
	"crypto/md5"
	"fmt"
	"sync"

	"sdle.com/mod/utils"
)

type NodeStatus int64

const (
	NODE_OK           NodeStatus = 0 // When a Node is responsive
	NODE_UNRESPONSIVE NodeStatus = 1 // When a Node is unresponsive
	NODE_UNKNOWN      NodeStatus = 2 // When a Node was recently added to the ring, and has never been communicated before
)

type NodeInfo struct {
	Id          string
	Address     string
	Port        string
	Status      NodeStatus
	GossipLock  sync.Mutex
	DeadCounter int64
}

/**
 * Creates the node information object
 */
func newNodeInfo(address string, port string, status NodeStatus) *NodeInfo {
	return &NodeInfo{
		Id:          fmt.Sprintf("%s:%s", address, port),
		Address:     address,
		Port:        port,
		Status:      status,
		DeadCounter: 0,
	}
}

type HashRing struct {
	vnodes *utils.AVLTree
	nodes  map[string]*NodeInfo
	lock   sync.Mutex
}

/**
 * Initialize the Hash Ring
 */
func (ring *HashRing) Initialize() {
	ring.vnodes = &utils.AVLTree{}
	ring.nodes = make(map[string]*NodeInfo)
}

/**
 * Adds a node to the hash ring
 */
func (ring *HashRing) AddNode(address string, port string, isServer bool) bool {
	ring.lock.Lock()

	result := ring.addNode(address, port, isServer)

	ring.lock.Unlock()

	return result
}

func (ring *HashRing) addNode(address string, port string, isServer bool) bool {
	if address == "" || port == "" {
		return false
	}

	// A node's id is made of a string of the address and the port
	var id string = fmt.Sprintf("%s:%s", address, port)

	// Checking if the node is already in the ring
	if ring.nodes[id] != nil {
		return false
	}

	for i := 0; i < 8; i++ { // TODO: Should the node inform how many it wants? It is hardcoded 8 as it is cassandra's choice
		var vnode_id string = fmt.Sprintf("%s_vnode%d", id, i)

		var vnode_hash string = hashId(vnode_id)

		ring.vnodes.Add(vnode_hash, id) // the Virtual Node's hash is the key, it then points to the node
	}

	// Add the nodeInfo to the ring
	if isServer {
		ring.nodes[id] = newNodeInfo(address, port, NODE_OK)
	} else {
		ring.nodes[id] = newNodeInfo(address, port, NODE_UNKNOWN)
	}

	return true
}

func (ring *HashRing) GetNodes() map[string]*NodeInfo {
	return ring.nodes
}

func (ring *HashRing) NodesGossip() map[string][]map[string]string {
	ring.lock.Lock()
	nodesOnTheRing := ring.GetNodes()

	nodesData := make(map[string][]map[string]string)
	nodesData["nodes"] = make([]map[string]string, len(nodesOnTheRing))

	for _, value := range ring.GetNodes() {
		nodesData["nodes"] = append(nodesData["nodes"], map[string]string{"address": value.Address, "port": value.Port, "status": fmt.Sprintf("%d", value.Status)})
	}

	ring.lock.Unlock()
	return nodesData
}

func (ring *HashRing) CheckForNewNodes(nodes []map[string]string, ownHostname string, ownPort string) {
	ring.lock.Lock()

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		if node["address"] == "" || node["port"] == "" {
			continue
		}

		if node["address"] == ownHostname && node["port"] == ownPort {
			continue
		}

		nodeId := fmt.Sprintf("%s:%s", node["address"], node["port"])

		if ring.nodes[nodeId] == nil {
			fmt.Printf("FOUND AN UNKNOWN NODE %s:%s \n", node["address"], node["port"])

			ring.addNode(node["address"], node["port"], false)
		}

		// FIXME: Do I do something with the status?
	}

	ring.lock.Unlock()
}

func (ring *HashRing) GetNodeForIdFromRing(id string) *NodeInfo {
	ring.lock.Lock()

	var hash_key string = hashId(id)

	avl_node := ring.vnodes.Search(hash_key)

	if avl_node == nil {
		ring.lock.Unlock()
		return nil
	}

	node_read := ring.nodes[avl_node.Value]

	ring.lock.Unlock()

	return node_read
}

/**
* Gets the N healthy node in the hash ring after a certain ID
 */
func (ring *HashRing) GetNHealthyNodesForID(id string, n int) []*NodeInfo {
	ring.lock.Lock()

	result := ring.getNHealthyNodesForID(id, n)

	ring.lock.Unlock()

	return result
}

func (ring *HashRing) getNHealthyNodesForID(id string, n int) []*NodeInfo {

	var hash_key string = hashId(id)

	if ring.vnodes.Size() < n {
		n = ring.vnodes.Size()
	}

	nodes := make([]*NodeInfo, 0)

	if n < 1 {
		return nodes
	}

	avlNode := ring.vnodes.Search(hash_key)

	if ring.nodes[avlNode.Value].Status == NODE_OK {
		nodes = append(nodes, ring.nodes[avlNode.Value])
	}

	// Get the current key so we can find the next
	hash_key = avlNode.GetKey()

	for i := 0; i < n-1; i++ {
		avlNode := ring.vnodes.Next(hash_key)
		if ring.nodes[avlNode.Value].Status == NODE_OK {
			nodes = append(nodes, ring.nodes[avlNode.Value])
		}

		// Get the current key so we can find the next
		hash_key = avlNode.GetKey()
	}

	// fmt.Println(nodes, ring.nodes)

	return nodes
}

func hashId(id string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(id)))
}
