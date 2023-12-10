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
}

// NewRoundRobinBalancer creates a new RoundRobinBalancer.
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{nextIndex: 0}
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
func (b *RoundRobinBalancer) SelectNodeFromList(list []*hash_ring.NodeInfo) *hash_ring.NodeInfo {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(list) == 0 {
		return nil
	}

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
