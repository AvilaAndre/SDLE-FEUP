package main

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"sdle.com/mod/protocol"
	"sdle.com/mod/utils"
)

var listsToRealocate utils.Stack[struct {
	listID       string
	correctNodes []string
}]

func hintedHandoff() {
	for {
		if listsToRealocate.Size() > 0 {

			// utils.Stack
			//Print the message I am doing hinted handoff
			log.Println("I am inside method  HINTED-OFF, listToRealocate:",  listsToRealocate,", Size of list: ",listsToRealocate.Size())


			//Print the lists that are going to be reallocated
			
			writeChan := make(chan struct {
				string
				bool
			})

			var waitingFor int = 0
			waitingForMap := make(map[string]int) // How many times do we need to wait for a successful write for each list

			for j := 0; j < listsToRealocate.Size(); j++ {
				listInfo := listsToRealocate.Pop()

				shoppingList, found := database.getShoppingList(listInfo.listID)
				if !found {
					continue
				}

				for k := 0; k < len(listInfo.correctNodes); k++ {
					parsedServerID := strings.FieldsFunc(listInfo.correctNodes[k], func(r rune) bool {
						return r == ':'
					})

					go sendHintedHandoffWriteAndWait(parsedServerID[0], parsedServerID[1], protocol.ShoppingListOperation{ListId: listInfo.listID, Content: shoppingList}, writeChan)
					waitingFor += 1
				}
				waitingForMap[listInfo.listID] = min(ring.ReplicationFactor/2+1, len(listInfo.correctNodes)) // same logic as the quorums
			}

			for {
				if waitingFor < 1 {
					break
				}
				result := <-writeChan

				waitingFor -= 1
				if result.bool {
					waitingForMap[result.string] -= 1
				}
			}

			for key, value := range waitingForMap {
				// If the list was reallocated successfully delete from this database
				if value <= 0 {
					database.deleteList(key)
				}
			}
		}
		log.Println("One ITERATION OF HINTED HANDOFF FINISHED")

		time.Sleep(10 * time.Second)
	}
}

// Returns true if successful, false if not
func sendHintedHandoffWriteAndWait(address string, port string, payload protocol.ShoppingListOperation, writeChan chan struct {
	string
	bool
}) {
	if address == serverHostname && port == serverPort {
		database.updateOrSetShoppingList(payload.ListId, payload.Content)

		writeChan <- struct {
			string
			bool
		}{payload.ListId, true}
		return
	}

	response, err := sendWrite(address, port, payload)
	if err != nil {
		writeChan <- struct {
			string
			bool
		}{payload.ListId, false}
		return
	}

	// Successful if write suceeds
	if response.StatusCode == http.StatusOK {
		writeChan <- struct {
			string
			bool
		}{payload.ListId, true}
	} else {
		writeChan <- struct {
			string
			bool
		}{payload.ListId, false}
	}
}

func checkForHintedHandoff() {
	// check for lists in the wrong partitions

	// Get lists
	listIds := make([]string, 0)

	database.lock.Lock()

	cursor, err := database.conn.NewCursor()

	if err != nil {
		cursor.Close()
		database.lock.Unlock()
		return
	}

	cursor.First()

	// iterate through every list key

	for {
		key, err := cursor.Key()

		if err != nil {
			break
		}
		listIds = append(listIds, string(key))

		cursor.Next()

	}

	cursor.Close()
	database.lock.Unlock()

	// For every list on the database check if they are on the right place

	selfAddress := fmt.Sprintf("%s:%s", serverHostname, serverPort)

	listsToRealocate.New()

	for i := 0; i < len(listIds); i++ {
		var listKey string = listIds[i]

		nextVNode := ring.GetNextHealthyVirtualNode(listKey)

		// If the list is in the wrong partition, append to the reallocation array and indicate the right nodes
		if !slices.Contains(ring.GetPartitions()[nextVNode], selfAddress) {
			listsToRealocate.Push(struct {
				listID       string
				correctNodes []string
			}{listKey, ring.GetPartitions()[nextVNode]})
		}
	}
}
