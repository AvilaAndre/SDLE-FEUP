package main

import (
	"fmt"
	"log"
	"net/http"

	"sdle.com/mod/protocol"
)

func handleList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		target := make(map[string]string)

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}

		fmt.Println("get", target)

		node_with_data := ring.GetNodeForId(target["list"])

		log.Println("node_with_data", node_with_data)

		log.Println(database.getValueRaw(target["list"]))

	/**
	 * This writes the data received into a key on the database
	 */
	case http.MethodPost:
		target := make(map[string]string) // TODO: Replace string with the CRDT

		decoded, target := protocol.DecodeRequestBody(w, r.Body, target)

		if !decoded {
			return
		}

		fmt.Println("post", target)

		database.writeToKey(target["list_id"], []byte(target["list"]))
	}
}
