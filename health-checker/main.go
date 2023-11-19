// this adds database-nodes and checks for their health

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
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

func pingNode(i int, index chan int, status chan string) {
	var nodeAddress string = nodes[i].address
		var nodePort string = nodes[i].port

		con, err := net.Dial("tcp", fmt.Sprintf("%s:%s", nodeAddress, nodePort))

		if (err != nil) {
			index <- i;
			status <- "UNRESPONSIVE"
			return
		}
		
		defer con.Close()
		
		msg := "ping"
		
		start := time.Now().UnixNano() / int64(time.Millisecond)

		_, err = con.Write([]byte(msg))
		
		checkErr(err)
		
		reply := make([]byte, 1024)
		
		size, err := con.Read(reply)
		
		checkErr(err)
		replyString := string(reply[:size])
		if (replyString == "pong") {
			end := time.Now().UnixNano() / int64(time.Millisecond)
			diff := end - start
			index <- i;
			status <- fmt.Sprintf("OK: %d ms", diff)
		}
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
	fmt.Printf("Address = %s\n", address)
	fmt.Printf("Port = %s\n", port)

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

