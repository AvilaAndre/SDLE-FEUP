// this adds database-nodes and checks for their health

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"sdle.com/mod/protocol"
)

type node struct {
    address string
    port string
    status string
}

func newNode(address string, port string) *node {
    p := node{address: address, port: port, status: "NOT YET CONNECTED"}
    return &p
}

type PingResponse struct {
	Message string
}

func pingNode(i int, index chan int, status chan string) {
	var nodeAddress string = nodes[i].address
	var nodePort string = nodes[i].port

	start := time.Now().UnixNano() / int64(time.Millisecond)
	response, err := protocol.SendGetRequest(nodeAddress, nodePort, "/ping")
	end := time.Now().UnixNano() / int64(time.Millisecond)
	diff := end - start

	if (err != nil) {
		index <- i;
		status <- "UNRESPONSIVE"
		return
	}

	if (response.StatusCode == 200) {
		target := PingResponse{} 
		
		json.NewDecoder(response.Body).Decode(&target)

		if (target.Message == "pong") {
			index <- i;
			status <- fmt.Sprintf("OK: %d ms", diff)
			return
		}
	}

	index <- i;
	status <- "ERROR"
}


func healthCheck() {
	indexChan := make(chan int)
	statusChan := make(chan string)

	for i := 0; i < len(nodes); i++ {
		go pingNode(i, indexChan, statusChan)
	}

	for i := 0; i < len(nodes); i++ {
		index, status := <- indexChan, statusChan
		nodes[index].status = <-status
	}
}

var responseHTML string = ""
var nodes []*node

func getRoot(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, responseHTML)
}

func getNodes(w http.ResponseWriter, r *http.Request) {
	healthCheck()

	var nodesInfo string = "";

	for i := 0; i < len(nodes); i++ {
		nodesInfo += "<h2>" + fmt.Sprintf("%s:%s - %s", nodes[i].address, nodes[i].port, nodes[i].status) + "</h2>\n";
	}

	io.WriteString(w, nodesInfo)
}

func getAdd(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /add request\n")

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	fmt.Printf("Post from website! r.PostFrom = %v\n", r.PostForm)
	address := r.FormValue("address")
	port := r.FormValue("port")

	nodes = append(nodes, newNode(address, port))

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	htmlContent, err1 := os.ReadFile("./health-checker/index.html")
    checkErr(err1)

	responseHTML = string(htmlContent)

	http.HandleFunc("/", getRoot)
	http.HandleFunc("/nodes", getNodes)
	http.HandleFunc("/add", getAdd)

	err := http.ListenAndServe(":3333", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func checkErr(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

