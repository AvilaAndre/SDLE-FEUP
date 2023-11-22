package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"sdle.com/mod/protocol"
	"sdle.com/mod/utils"
)

func startServerAndJoinCluster(serverPort string, loadBalancerAddress string, loadBalancerPort string, ownData map[string]string) {
	serverRunning := make(chan bool)

	go startServer(serverPort, serverRunning)

	// FIXME: Should I wait for the server to start?

	go joinCluster(loadBalancerAddress, loadBalancerPort, ownData)

	<-serverRunning // waits for the server to close
}

func startServer(serverPort string, serverRunning chan bool) {
	err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed")
	} else if err != nil {
		log.Printf("error starting server: %s\n", err)
	}

	serverRunning <- true
}

func joinCluster(loadBalancerAddress string, loadBalancerPort string, ownData map[string]string) {
	if loadBalancerAddress != "" && loadBalancerPort != "" {
		r, err := protocol.SendRequestWithData(http.MethodPut, loadBalancerAddress, loadBalancerPort, "/node/add", ownData)
		utils.CheckErr(err)

		if r.StatusCode == 202 {
			log.Println("Joined the cluster")
		} else {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("Failed to join the cluster")
			} else {
				log.Println("Failed to join the cluster:", string(body))
			}
			os.Exit(1)
			return
		}

		target := make(map[string][]map[string]string)

		err = json.NewDecoder(r.Body).Decode(&target)
		utils.CheckErr(err)

		for i := 0; i < len(target["nodes"]); i++ {
			newNode := target["nodes"][i]

			nodes.addNode(node{address: newNode["address"], port: newNode["port"]})
		}
	}

}
