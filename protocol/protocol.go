package protocol

import (
	"fmt"
	"net/http"
	"time"
)


func SendGetRequest(address string, port string, path string) (*http.Response, error) {
	requestURL := fmt.Sprintf("http://%s:%s%s", address, port, path)
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Get(requestURL)
	if err != nil {
		return nil, err
	}
	
	return res, nil
}