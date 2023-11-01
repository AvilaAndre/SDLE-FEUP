package main

import (
	"github.com/zeromq/goczmq"
)

type DatabaseNode struct {
    id int
	endpoint string
	port string
	sock *goczmq.Sock
}

func NewDatabaseNode(id int, endpoint string, port string, sock *goczmq.Sock) (*DatabaseNode) {
	return &DatabaseNode{id: id, endpoint: endpoint, port: port, sock: sock}
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