package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sdle.com/mod/hash_ring"
	"sdle.com/mod/utils"
	"sync"
	"time"
)

var serverPort string = ""
var serverHostname string = ""
var ring hash_ring.HashRing
var roundRobinBalancer = NewRoundRobinBalancer()

// Node represents information about a node.
type Node struct {
	add  string
	port string
}

// RoundRobinBalancer is a simple round-robin load balancer for nodes.
type RoundRobinBalancer struct {
	nodes     []*Node
	nextIndex int
	lock      sync.Mutex
	requestCount (map[string]int)
}

// NewRoundRobinBalancer creates a new RoundRobinBalancer.
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{nextIndex: 0,
		requestCount: make(map[string]int)}
}
// IncrementRequestCount increases the request count for a given node.
func (b *RoundRobinBalancer) IncrementRequestCount(nodeId string) {
    b.lock.Lock()
	// print inside the the IncrementRequestCount
	fmt.Println("Inside IncrementRequestCount for node: ", nodeId)
	// if the node is not in the map, add it
	if _, ok := b.requestCount[nodeId]; !ok {
		b.requestCount[nodeId] = 0
	}else{

		b.requestCount[nodeId]++
	}
	fmt.Println("Inside IncrementRequestCount: requestCount of load balancer: ", b.requestCount,"for node: ", nodeId)
	b.lock.Unlock()
}
// AddNode adds a new node to the balancer.
func (b *RoundRobinBalancer) AddNode(add string, port string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	node := &Node{add: add, port: port}
	b.nodes = append(b.nodes, node)
}

// RemoveNode removes a node from the balancer.
func (b *RoundRobinBalancer) RemoveNode(add string, port string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	var updatedNodes []*Node
	for _, node := range b.nodes {
		if node.add != add && node.port != port {
			updatedNodes = append(updatedNodes, node)
		}
	}
	b.nodes = updatedNodes
}

// SelectNode returns the next node in a round-robin fashion.
func (b *RoundRobinBalancer) SelectNode() *Node {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.nodes) == 0 {
		return nil
	}

	node := b.nodes[b.nextIndex]
	b.nextIndex = (b.nextIndex + 1) % len(b.nodes)
	return node
}

// SelectNodeFromList returns the next node in a round-robin fashion but the solution must be one of the input nodes.
func (b *RoundRobinBalancer) SelectNodeFromList(list []*hash_ring.NodeInfo , threshold float64, cons_hash_req_count_node int) *hash_ring.NodeInfo {
	b.lock.Lock()
	defer b.lock.Unlock()
	 

	if len(list) == 0 {
		return nil
	}
	var leastLoadedNode *hash_ring.NodeInfo
    var mostLoadedNode *hash_ring.NodeInfo
    minRequestCount := int(^uint(0) >> 1) // Max int value
    maxRequestCount := 0
	
	fmt.Println("len of list: ", len(list))
	if(len(list) == 1){
		return list[0]
	}
	// print  the map of requestCount for each node
	fmt.Println("requestCount of load balancer: ", b.requestCount)

    for _, node := range list {
        if count, ok := b.requestCount[node.Id]; ok {
            if count <= minRequestCount {
                minRequestCount = count
                leastLoadedNode = node
            }
            if count > maxRequestCount {
                maxRequestCount = count
                mostLoadedNode = node
            }
        }
    }

	// Select a node based on the load balancing criteria and the threshold
	// bounded consistent hashing to avoid node overload for a single list_ids !!
	// print the consistentHashNode.Id 
	fmt.Println("Getted consistentHashNode: ", list[0])	
	fmt.Println("maxRequestCount: ", maxRequestCount, " minRequestCount: ", minRequestCount, " threshold: ", threshold)
	fmt.Println("mostLoadedNode: ", mostLoadedNode, " leastLoadedNode: ", leastLoadedNode.Id)

	
	fmt.Println("consistentHashNodeCount: ", cons_hash_req_count_node)

    if mostLoadedNode != nil || leastLoadedNode != nil {
		if cons_hash_req_count_node > 100 && float64(cons_hash_req_count_node) > float64(minRequestCount)*(1.0+float64(threshold)){ 
            // If the most loaded node is overloaded ( using threshold), select the least loaded node
			fmt.Println("Selected the least loaded node")
            b.IncrementRequestCount(leastLoadedNode.Id)
		
            return leastLoadedNode
        }

	
	}
	fmt.Println("Selected the consistentHashNode")
	fmt.Println("requestCount of load balancer: ", b.requestCount[list[0].Id],"for node: ", list[0].Id)
	
	return list[0]
}

func (b *RoundRobinBalancer) SelectNodeFromListForAddNode(list []*hash_ring.NodeInfo) *hash_ring.NodeInfo {
	rand.Seed(time.Now().UnixNano())//enforce random seed for each request to avoid same node selection "random pattern" for each request
	
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(list) == 0 {
		return nil
	}
	// round robin approach for add nodes with load balancer
	node := list[rand.Intn(len(list))]
	return node
}



func main() {
	
	argsWithoutProg := os.Args[1:]
	
	if len(argsWithoutProg) > 0 {
		serverPort = argsWithoutProg[0]
		serverHostname = utils.GetOutboundIP().String()
	}

	if serverPort == "" {
		fmt.Println("A server port must be specified")
		os.Exit(1)
	}
	ring.Initialize()
	log.Printf("Load Balancer %s:%s", serverHostname, serverPort)
	

	serverRunning := make(chan bool)
	go func() {
		ticker := time.Tick(1 * time.Second)
		for {
			select {
			case <-ticker:
				gossip()
			case <-serverRunning:
				fmt.Println("Server is no longer running. Exiting.")
				return
			}
		}
	}()
	startServer(serverRunning)
	<-serverRunning // waits for the server to close
}
