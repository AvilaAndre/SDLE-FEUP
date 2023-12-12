package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"sdle.com/mod/hash_ring"
	"sdle.com/mod/protocol"
)

var threshold = 0.4 // Threshold for bounded consistent hashing
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
			// Print on Function routeCoordenator MethodPost
			fmt.Println("routeCoordenator MethodPost")
			target := make(map[string]string)

			buf, _ := io.ReadAll(request.Body)
			rdr1 := io.NopCloser(bytes.NewBuffer(buf))
			rdr2 := io.NopCloser(bytes.NewBuffer(buf))

			request.Body = rdr2

			decoded, target := protocol.DecodeRequestBody(writer, rdr1, target)

			if !decoded {
				return
			}

			var listId string = target["list_id"]

			// the healthyNodes are the nodes that are alive and they by order on the healthyNodes array from the ring
			
			healthyNodes := ring.GetHealthyNodesForID(listId)
			// so we get the first node from the healthyNodes array ( the first on encounter on the ring for list_id)
			if len(healthyNodes) == 0 {
				writer.WriteHeader(http.StatusNotFound)
				return
			}
			consistentHashNode := healthyNodes[0]

			fmt.Println("list_id to get the node using consistent hashing: ", listId)
			
			fmt.Println("Getted consistentHashNode: ", consistentHashNode)
			//print the healthyNodes
			fmt.Println("Getted healthyNodes: ", healthyNodes)
			fmt.Println("Getted consistentHashNode: ", healthyNodes[0])
			//print the healthyNodes
			// enter the bounded consistent hashing selection on SelectNodeFromList
			
			cons_hash_req_count_node := roundRobinBalancer.requestCount[healthyNodes[0].Id]
			// print the request count of the node
			fmt.Println("request count of node: ", cons_hash_req_count_node,"from node: ", healthyNodes[0].Id)

			node := roundRobinBalancer.SelectNodeFromList(healthyNodes,threshold,cons_hash_req_count_node)
			roundRobinBalancer.IncrementRequestCount(node.Id)
			proxy := httputil.NewSingleHostReverseProxy(&url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%s", node.Address, node.Port),
			})
			
			proxy.ServeHTTP(writer, request)
			return
		}

	case http.MethodPut:
		{	
			// Print on Function routeOperation MethodPut
			fmt.Println("routeCoordenator MethodPut")
			var target protocol.ShoppingListOperation

			buf, _ := io.ReadAll(request.Body)
			rdr1 := io.NopCloser(bytes.NewBuffer(buf))
			rdr2 := io.NopCloser(bytes.NewBuffer(buf))

			request.Body = rdr2

			decoded, target := protocol.DecodeRequestBody(writer, rdr1, target)

			if !decoded {
				return
			}
			
			fmt.Println("list_id to get the node using consistent hashing: ", target.ListId)
			// the healthyNodes are the nodes that are alive and they by order on the healthyNodes array from the ring
			
			healthyNodes := ring.GetHealthyNodesForID(target.ListId)
			// print len of healthyNodes
			fmt.Println("len of healthyNodes: ", len(healthyNodes))
			if len(healthyNodes) == 0 {
				// print if the len of healthyNodes is 0 then return 404
				fmt.Println("len of healthyNodes is 0 then return 404")
				writer.WriteHeader(http.StatusNotFound)
				return
			}
			// so we get the first node from the healthyNodes array ( the first on encounter on the ring for list_id)
			
			

			// print the list_id to determine the node using consistent hashing
			fmt.Println("list_id to get the node using consistent hashing: ", target.ListId)
			
			fmt.Println("Getted healthyNodes: ", healthyNodes)
			fmt.Println("Getted consistentHashNode: ", healthyNodes[0])
			//print the healthyNodes
			// enter the bounded consistent hashing selection on SelectNodeFromList
			
			cons_hash_req_count_node := roundRobinBalancer.requestCount[healthyNodes[0].Id]
			// print the request count of the node
			fmt.Println("request count of node: ", cons_hash_req_count_node,"from node: ", healthyNodes[0].Id)

			node := roundRobinBalancer.SelectNodeFromList(healthyNodes,threshold,cons_hash_req_count_node)
			// print the node selected
			fmt.Println("Chosen Getted inside Coordenator Method Put node: ", node)
			//print node address and port
			fmt.Println("Chosen Getted inside Coordenator Method Put node address: ", node.Address, " and port: ", node.Port)
			roundRobinBalancer.IncrementRequestCount(node.Id)
			proxy := httputil.NewSingleHostReverseProxy(&url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%s", node.Address, node.Port),
			})
			
			proxy.ServeHTTP(writer, request)
			return
		}

	}
}

func routeOperation(writer http.ResponseWriter, request *http.Request) {
	rand.Seed(time.Now().UnixNano())//enforce random seed for each request to avoid same node selection "random pattern" for each request
	
	switch request.Method {
	case http.MethodPost:
		{	

			// Print on Function routeOperation MethodPost
			fmt.Println("routeOperation MethodPost")
			target := make(map[string]string)
			decoded, target := protocol.DecodeRequestBody(writer, request.Body, target)

			if !decoded {
				return
			}

			var listId string = target["list_id"]

			
			// the healthyNodes are the nodes that are alive and they by order on the healthyNodes array from the ring
			healthyNodes := ring.GetHealthyNodesForID(listId)
			if len(healthyNodes) == 0 {
				writer.WriteHeader(http.StatusNotFound)
				return
			}
			// so we get the first node from the healthyNodes array ( the first on encounter on the ring for list_id)
			
			
			
			fmt.Println("list_id to get the node using consistent hashing: ", listId)
			
			fmt.Println("Getted consistentHashNode: ", healthyNodes[0])
			//print the healthyNodes
			// enter the bounded consistent hashing selection on SelectNodeFromList
			
			cons_hash_req_count_node := roundRobinBalancer.requestCount[healthyNodes[0].Id]
			// print the request count of the node
			fmt.Println("request count of node: ", cons_hash_req_count_node,"from node: ", healthyNodes[0].Id)

			node := roundRobinBalancer.SelectNodeFromList(healthyNodes,threshold,cons_hash_req_count_node)
			
			roundRobinBalancer.IncrementRequestCount(node.Id)
			proxy := httputil.NewSingleHostReverseProxy(&url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%s", node.Address, node.Port),
			})
			
			proxy.ServeHTTP(writer, request)
			// print the increment request count of the node
			fmt.Println("request count of node: ", roundRobinBalancer.requestCount[node.Id],"from node: ", node.Id)

			return

		}
	case http.MethodPut:
		{
			// Print on Function routeOperation MethodPut
			fmt.Println("routeOperation MethodPut")

			var target protocol.ShoppingListOperation

			decoded, target := protocol.DecodeRequestBody(writer, request.Body, target)

			if !decoded {
				return
			}

			

			
			// the healthyNodes are the nodes that are alive and they by order on the healthyNodes array from the ring
			healthyNodes := ring.GetHealthyNodesForID(target.ListId)
			if len(healthyNodes) == 0 {
				writer.WriteHeader(http.StatusNotFound)
				return
			}
			// so we get the first node from the healthyNodes array ( the first on encounter on the ring for list_id)
			
			consistentHashNode := healthyNodes[0]
			
			
			fmt.Println("list_id to get the node using consistent hashing: ", target.ListId)
			
			fmt.Println("Getted consistentHashNode: ", consistentHashNode)
			fmt.Println("Getted consistentHashNode: ", healthyNodes[0])
			//print the healthyNodes
			// enter the bounded consistent hashing selection on SelectNodeFromList
			
			cons_hash_req_count_node := roundRobinBalancer.requestCount[healthyNodes[0].Id]
			// print the request count of the node
			fmt.Println("request count of node: ", cons_hash_req_count_node,"from node: ", healthyNodes[0].Id)

			node := roundRobinBalancer.SelectNodeFromList(healthyNodes,threshold,cons_hash_req_count_node)
			roundRobinBalancer.IncrementRequestCount(node.Id)
			proxy := httputil.NewSingleHostReverseProxy(&url.URL{
				Scheme: "http",
				Host:   fmt.Sprintf("%s:%s", node.Address, node.Port),
			})
			
			proxy.ServeHTTP(writer, request)
			// print the increment request count of the node
			fmt.Println("request count of node: ", roundRobinBalancer.requestCount[node.Id],"from node: ", node.Id)

			return
		}
	}
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
			// adding node initially in map of request count
			roundRobinBalancer.IncrementRequestCount(target["address"]+":"+target["port"])

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
	node := roundRobinBalancer.SelectNodeFromListForAddNode(nodes)

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


