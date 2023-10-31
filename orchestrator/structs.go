package main

import (
	"github.com/zeromq/goczmq"
)

type DatabaseNode struct {
    Id int
	Sock *goczmq.Sock
}