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

		if !protocol.DecodeRequestBody(w, r.Body, target) {
			return
		}

		log.Println(database.getValueRaw(target["list"]))

	/**
	 * This writes the data received into a key on the database
	 */
	case http.MethodPost:
		target := make(map[string]string) // TODO: Replace string with the CRDT

		if !protocol.DecodeRequestBody(w, r.Body, target) {
			return
		}

		fmt.Println("post", target)

		database.writeToKey(target["list_id"], []byte(target["list"]))
	}
}
