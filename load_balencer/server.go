package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sdle.com/mod/hash_ring"
	"sdle.com/mod/protocol"
	"time"
)

func registerRoutes() {
	http.HandleFunc("/operation", routeOperation)
	http.HandleFunc("/list", routeCoordenator)
	http.HandleFunc("/node/add", addNode)
	http.HandleFunc("/ping", Ping)
}

func startServer(serverRunning chan bool) {
	registerRoutes()
	err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed")
	} else if err != nil {
		log.Printf("error starting server: %s\n", err)
	}

	serverRunning <- true
}

func routeCoordenator(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		{
			target := make(map[string]string)
			decoded, target := protocol.DecodeRequestBody(writer, request.Body, target)

			if !decoded {
				return
			}

			var listId string = target["list_id"]

			// The coordenator, upon receiving a read, reads locally and performs a read quorum
			// however, this coordenator may not be a holder of this information, in this case
			// it only performs the read quorum
			healthyNodes := ring.GetHealthyNodesForID(listId)

			node := roundRobinBalancer.SelectNodeFromList(healthyNodes)

			// proxy the request to the selected node
			body, _ := json.Marshal(target)
			data, err := protocol.SendRequestWithData(http.MethodPost, node.Address, node.Port, "/operation", body)
			if err != nil {
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(data.StatusCode)

			//send data back to client
			io.Copy(writer, data.Body)

			return
		}

	}
}

func routeOperation(writer http.ResponseWriter, request *http.Request) {
	target := make(map[string]string)
	decoded, target := protocol.DecodeRequestBody(writer, request.Body, target)

	if !decoded {
		return
	}

	var listId string = target["list_id"]

	// The coordenator, upon receiving a read, reads locally and performs a read quorum
	// however, this coordenator may not be a holder of this information, in this case
	// it only performs the read quorum
	healthyNodes := ring.GetHealthyNodesForID(listId)

	node := roundRobinBalancer.SelectNodeFromList(healthyNodes)

	// proxy the request to the selected node
	body, _ := json.Marshal(target)
	data, err := protocol.SendRequestWithData(http.MethodPost, node.Address, node.Port, "/operation", body)
	if err != nil {
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(data.StatusCode)

	//send data back to client
	io.Copy(writer, data.Body)

	return
}

func gossip() {
	for {
		for _, value := range ring.GetNodes() {
			go gossipWith(value)
		}
		time.Sleep(1 * time.Second) // FIXME: This should not be like this :p
	}
}

func gossipWith(node *hash_ring.NodeInfo) {
	if !node.GossipLock.TryLock() {
		return
	}

	gossipMaterial := ring.NodesGossip()

	jsonData, err := json.Marshal(gossipMaterial)
	if err != nil {
		log.Printf("Error happened in JSON marshal. Err: %s", err)
		node.GossipLock.Unlock()
		return
	}

	response, err2 := protocol.SendRequestWithData(http.MethodPost, node.Address, node.Port, "/gossip", jsonData)
	if err2 != nil {
		// If cannot gossip four consecutive times then assume the node is dead
		if node.DeadCounter < 3 {
			node.DeadCounter++
		} else if node.DeadCounter == 3 {
			node.Status = hash_ring.NODE_UNRESPONSIVE
			log.Printf("%s set to UNRESPONSIVE\n", node.Id)
			node.DeadCounter++
		}

		node.GossipLock.Unlock()
		return
	}

	// Await for response to get the node's status
	if response.StatusCode == 200 {
		node.DeadCounter = 0
		if node.Status != hash_ring.NODE_OK {
			node.Status = hash_ring.NODE_OK
			log.Printf("%s set to OK\n", node.Id)
		}
	}

	node.GossipLock.Unlock()
}

func Ping(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("pong"))
}

func addNode(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPut:
		{
			target := make(map[string]string)

			decoded, target := protocol.DecodeRequestBody(writer, request.Body, target)

			if !decoded {
				return
			}

			var isServer bool = target["address"] == serverHostname && target["port"] == serverPort

			ring.AddNode(target["address"], target["port"], isServer)
			jsonData := encodeRingState()

			roundRobinBalancer.AddNode(target["address"], target["port"])

			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusAccepted)
			writer.Write(jsonData)
			sendToRoundRobinBalancer(target["address"], target["port"])
			return
		}
	}
}

func encodeRingState() []byte {
	nodesOnTheRing := ring.GetNodes()

	nodesData := make(map[string][]map[string]string)
	nodesData["nodes"] = make([]map[string]string, len(nodesOnTheRing))

	for _, value := range ring.GetNodes() {
		nodesData["nodes"] = append(nodesData["nodes"], map[string]string{"address": value.Address, "port": value.Port})
	}

	jsonData, err := json.Marshal(nodesData)

	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	return jsonData
}

func sendToRoundRobinBalancer(address string, port string) {
	var1 := ring.GetNodes()
	nodes := make([]*hash_ring.NodeInfo, 0)
	for _, value := range var1 {
		if value.Status != hash_ring.NODE_UNRESPONSIVE {
			nodes = append(nodes, value)
		}
	}
	node := roundRobinBalancer.SelectNodeFromList(nodes)

	target := make(map[string]string)
	target["address"] = address
	target["port"] = port
	j, _ := json.Marshal(target)

	data, err := protocol.SendRequestWithData(http.MethodPut, node.Address, node.Port, "/node/add", j)
	if err != nil {
		return
	}

	if data.StatusCode == 202 {
		log.Println("Rings are in sync")
	} else {
		body, err := io.ReadAll(data.Body)
		if err != nil {
			log.Println("Failed to sync the rings")
		} else {
			log.Println("Failed to sync the rings:", string(body))
		}
		return
	}
}
