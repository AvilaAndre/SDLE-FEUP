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

func (lb *LoadBalencer) Start() {
	serverRunning := make(chan bool)
	lb.startServer(lb.port, serverRunning)
	<-serverRunning // waits for the server to close
}

func (lb *LoadBalencer) startServer(serverPort string, serverRunning chan bool) {


	err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("server closed")
	} else if err != nil {
		log.Printf("error starting server: %s\n", err)
	}

	serverRunning <- true
	
}
