package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

func SendRequestWithData(method string, address string, port string, path string, data map[string]string) (*http.Response, error) {
	requestURL := fmt.Sprintf("http://%s:%s%s", address, port, path)

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		return nil, err
	}

	req, err := http.NewRequest(method, requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	return res, nil
}
