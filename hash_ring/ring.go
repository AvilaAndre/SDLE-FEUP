package hash_ring
// TODO: in practice we use should use k ≈ log2(num_items) to obtain reasonably balanced loads for storing a total of num_items in the entire datastore
import (
	"crypto/md5"
	"fmt"
	"strings"
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
	Vnodes      []string
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
		Vnodes:      make([]string, 0),
		DeadCounter: 0,
	}
}

func (nI *NodeInfo) GetVirtualNodes() []string {
	return nI.vnodes
}

type HashRing struct {
	vnodes            *utils.AVLTree
	nodes             map[string]*NodeInfo
	ReplicationFactor int
	partitions        map[string][]string
	updated           bool
	lock              sync.Mutex
}

/**
 * Initialize the Hash Ring
 */
func (ring *HashRing) Initialize() {
	ring.vnodes = &utils.AVLTree{}
	ring.nodes = make(map[string]*NodeInfo)
	ring.partitions = make(map[string][]string)
	ring.ReplicationFactor = 1
}

/**
 * Gets the hash ring's partitions
 */
func (ring *HashRing) GetPartitions() map[string][]string {
	return ring.partitions
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

	// Add the nodeInfo to the ring
	if isServer {
		ring.nodes[id] = newNodeInfo(address, port, NODE_OK)
	} else {
		ring.nodes[id] = newNodeInfo(address, port, NODE_UNKNOWN)
	}

	// Update the hash ring
	ring.updateRing()

	return true
}

func (ring *HashRing) addVirtualNode(node_id string, vnode_number int) {
	var vnode_id string = fmt.Sprintf("%s_vnode%d", node_id, vnode_number)

	var vnode_hash string = HashId(vnode_id)

	ring.vnodes.Add(vnode_hash, vnode_id) // the Virtual Node's hash is the key, it then points to the node

	// Add vnode_id to the vnodes list
	ring.nodes[node_id].Vnodes = append(ring.nodes[node_id].Vnodes, vnode_id)
}

// TODO: This is only called if a node is deleted
func (ring *HashRing) removeVirtualNode(node_id string, vnode_number int) {
	var vnode_id string = fmt.Sprintf("%s_vnode%d", node_id, vnode_number)

	var vnode_hash string = HashId(vnode_id)

	ring.vnodes.Remove(vnode_hash) // the Virtual Node's hash is the key, it then points to the node

	// Add vnode_id to the vnodes list
	for index := 0; index < len(ring.nodes[node_id].Vnodes); index++ {
		if ring.nodes[node_id].Vnodes[index] == vnode_id {
			ring.nodes[node_id].Vnodes = append(ring.nodes[node_id].Vnodes[:index], ring.nodes[node_id].Vnodes[index+1:]...)
			break
		}
	}
}

func (ring *HashRing) updateRing() {
	// determine maximum number of nodes
	var max_vnodes int = min(8, len(ring.nodes)-1)

	if max_vnodes == 0 {
		max_vnodes = 1
	}

	ring.ReplicationFactor = max_vnodes

	// check for every node if it has the correct ammount of vnodes

	for node_id := range ring.nodes {
		var number_of_vnodes int = len(ring.nodes[node_id].Vnodes)
		if number_of_vnodes < max_vnodes {
			// Adds the missing Virtual Nodes from this node
			for i := number_of_vnodes; i < max_vnodes; i++ {
				ring.addVirtualNode(node_id, i)
			}
		} else if number_of_vnodes > max_vnodes {
			// Removes the excess of Virtual Nodes from this node
			for i := number_of_vnodes - 1; i >= max_vnodes; i-- {
				ring.removeVirtualNode(node_id, i)
			}
		}
	}

	// update partitions

	ring.partitions = make(map[string][]string) // map of vnodes and what nodes they write into
	vnodes := make([]string, 0)

	for _, v := range ring.GetNodes() {
		vnodes = append(vnodes, v.GetVirtualNodes()...)
	}

	// Get what nodes have this partition

	for i := 0; i < len(vnodes); i++ {
		vnodeHash := fmt.Sprintf("%x", md5.Sum([]byte(vnodes[i])))

		ring.partitions[vnodeHash] = make([]string, 0)

		rawHealthynodes := ring.getHealthyNodesForID(vnodeHash)
		healthyNodes := rawHealthynodes[:min(ring.ReplicationFactor, len(rawHealthynodes))]

		for j := 0; j < len(healthyNodes); j++ {
			// if node is in the healthy nodes then he stores this partition
			ring.partitions[vnodes[i]] = append(ring.partitions[vnodes[i]], healthyNodes[j].Id)
		}
	}

	// already locked, no need for locking
	ring.updated = true
}

/**
* Sets the ring as updated
 */
func (ring *HashRing) setUpdated() {
	ring.lock.Lock()
	ring.updated = true
	ring.lock.Unlock()
}

/**
* Indicates if the hash ring was recently updated
 */
func (ring *HashRing) WasUpdated() bool {
	ring.lock.Lock()

	wasUpdated := ring.updated
	ring.updated = false

	ring.lock.Unlock()

	return wasUpdated
}

func (ring *HashRing) NodeStatusChanged() {
	ring.lock.Lock()
	ring.updateRing()
	ring.lock.Unlock()
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

	}

	ring.lock.Unlock()
}

func (ring *HashRing) GetNodeForIdFromRing(id string) *NodeInfo {
	ring.lock.Lock()

	var hash_key string = HashId(id)

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
func (ring *HashRing) GetHealthyNodesForID(id string) []*NodeInfo {
	ring.lock.Lock()

	result := ring.getHealthyNodesForID(id)

	ring.lock.Unlock()

	return result
}

func (ring *HashRing) GetNextHealthyVirtualNode(id string) string {
	var hash_key string = hashId(id)

	avlNode := ring.vnodes.Search(hash_key)

	var firstVNodeId string = avlNode.Value

	// parse virtual node name <node_name>_vnode<id>
	parsedServerName := ring.ParseVirtualNodeID(avlNode.Value)[0]

	if ring.nodes[parsedServerName].Status == NODE_OK {
		return firstVNodeId
	}

	// Get the current key so we can find the next
	hash_key = avlNode.GetKey()

	for {
		avlNode := ring.vnodes.Next(hash_key)
		// Get the current key so we can find the next
		hash_key = avlNode.GetKey()

		// If the node is the first then it is looping
		if avlNode.Value == firstVNodeId {
			break
		}

		parsedServerName := ring.ParseVirtualNodeID(avlNode.Value)[0]

		if ring.nodes[parsedServerName].Status == NODE_OK {
			return avlNode.Value
		}
	}

	// Only reaches this if every node is unresponsive
	return firstVNodeId
}

// Calculates the n nodes after an id and returns them
func (ring *HashRing) getHealthyNodesForID(id string) []*NodeInfo {
	nodes := make([]*NodeInfo, 0)

	if ring.vnodes.Size() < 1 {
		return nodes
	}

	var hash_key string = HashId(id)

	nodesChecked := make([]string, 0)

	avlNode := ring.vnodes.Search(hash_key)

	var firstVNodeId string = avlNode.Value

	// parse virtual node name <node_name>_vnode<id>
	parsedServerName := ring.ParseVirtualNodeID(avlNode.Value)[0]

	if ring.nodes[parsedServerName].Status == NODE_OK {
		nodes = append(nodes, ring.nodes[parsedServerName])
		nodesChecked = append(nodesChecked, parsedServerName)
	}

	// Get the current key so we can find the next
	hash_key = avlNode.GetKey()

	for {
		avlNode := ring.vnodes.Next(hash_key)
		// Get the current key so we can find the next
		hash_key = avlNode.GetKey()

		// If the node is the first then it is looping
		if avlNode.Value == firstVNodeId {
			break
		}

		parsedServerName := ring.ParseVirtualNodeID(avlNode.Value)[0]

		// Check if this node has already been found in the ring
		var isRepeated bool = false
		for i := 0; i < len(nodesChecked); i++ {
			if nodesChecked[i] == parsedServerName {
				isRepeated = true
				break
			}
		}

		if isRepeated {
			continue
		} else {
			nodesChecked = append(nodesChecked, parsedServerName)
		}

		if ring.nodes[parsedServerName].Status == NODE_OK {
			nodes = append(nodes, ring.nodes[parsedServerName])
		}

		// if the all nodes have been found in the ring, break
		if len(nodesChecked) == len(ring.nodes) {
			break
		}
	}

	return nodes
}

func HashId(id string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(id)))
}

func (ring *HashRing) ParseVirtualNodeID(virtualNodeID string) []string {
	return strings.FieldsFunc(virtualNodeID, func(r rune) bool {
		return r == '_'
	})
}
