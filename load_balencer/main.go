package load_balencer

import (
)

type LoadBalencer struct {
	address string
	port string
}

func (lb *LoadBalencer) Initialize(address string, port string) {
	lb.address = address
	lb.port = port
}