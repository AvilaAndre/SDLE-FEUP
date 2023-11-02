package main

import (
	"fmt"

	"github.com/zeromq/goczmq"
	"sdle.com/mod/protocol"
	utils "sdle.com/mod/utils"
)

type DatabaseNode struct {
    id int
	endpoint string
	port string
	sock *goczmq.Sock
	left_node int
	right_node int
}

func NewDatabaseNode(endpoint string, port string, sock *goczmq.Sock) (*DatabaseNode) {
	return &DatabaseNode{id: -1, endpoint: endpoint, port: port, sock: sock, left_node: -1, right_node: -1}
}

func (db_node *DatabaseNode) GetId() int {
	return db_node.id
}

func (db_node *DatabaseNode) GetEndpoint() string {
	return db_node.endpoint
}

func (db_node *DatabaseNode) GetPort() string {
	return db_node.port
}

func (db_node *DatabaseNode) GetSock() (*goczmq.Sock) {
	return db_node.sock
}

type DatabaseClusterOrganization struct {
	nodes map[int]*DatabaseNode
	last_id int
	first_id int
}

func NewDatabaseClusterOrganization() *DatabaseClusterOrganization {
	return &DatabaseClusterOrganization{nodes: make(map[int]*DatabaseNode), last_id: -1, first_id: -1}
}

func (cluster *DatabaseClusterOrganization) AddNewDatabaseNode(new_node *DatabaseNode) {
	// Get new ID for the node
	var new_id int = cluster.last_id+1
	// Adding the node to the cluster
	cluster.nodes[new_id] = new_node

	// Update first and last nodes' neighbours
	// And notify modified nodes
	if cluster.last_id != -1 {
		cluster.nodes[cluster.last_id].right_node = new_id
		utils.SendMessageAndWaitAck(cluster.nodes[cluster.last_id].sock, protocol.UpdateNodeRightNeighbourMessage(new_node.endpoint, new_node.port))
	}
	if cluster.first_id != -1 {
		cluster.nodes[cluster.first_id].left_node = new_id
		utils.SendMessageAndWaitAck(cluster.nodes[cluster.first_id].sock, protocol.UpdateNodeLeftNeighbourMessage(new_node.endpoint, new_node.port))
	}
	
	if cluster.first_id == -1 {
		cluster.first_id = new_id
	}
	
	// Update new last_id
	cluster.last_id = new_id
	utils.SendMessageAndWaitAck(new_node.sock, protocol.UpdateNodeIDMessage(fmt.Sprintf("%d", new_id)))
}
