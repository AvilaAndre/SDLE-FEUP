package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func SendRequestWithData(method string, address string, port string, path string, data []byte) (*http.Response, error) {
	requestURL := fmt.Sprintf("http://%s:%s%s", address, port, path)

	req, err := http.NewRequest(method, requestURL, bytes.NewBuffer(data))
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

const (
	JSON_DECODE_ERROR string = "Failed to decode the given JSON."
)

func FailedToDecodeJSON(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(JSON_DECODE_ERROR))
}

func RequestWithWrongFormat(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("This request is in the wrong format."))
}

func WrongRequestType(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Wrong protocol."))
}

/**
* Returns false if it fails to decode the body into the requested data format
 */
func DecodeRequestBody(w http.ResponseWriter, body io.ReadCloser, data any) bool {
	err := json.NewDecoder(body).Decode(&data)

	if err != nil {
		FailedToDecodeJSON(w)
		return false
	} else {
		return true
	}
}
