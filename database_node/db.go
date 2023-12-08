package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/nobonobo/unqlitego"
	"sdle.com/mod/crdt_go"
	"sdle.com/mod/utils"
)

type DatabaseInstance struct {
	conn *unqlitego.Database
	lock sync.Mutex
}

func (db *DatabaseInstance) initialize(address string, port string) {
	os.MkdirAll("./db", os.ModePerm)

	dbPath := fmt.Sprintf("./db/unqlite-%s:%s.db", address, port)

	conn, err := unqlitego.NewDatabase(dbPath)

	utils.CheckErr(err)

	db.conn = conn

	_, startErr := db.conn.Fetch([]byte("start"))

	// It means the database wasn't started
	if startErr.Error() == "IO error" {
		log.Panicln("Failed to create the database")
	}

	err2 := db.conn.Commit()

	utils.CheckErr(err2)

	log.Println("Database initialized at", dbPath)

}

func (db *DatabaseInstance) updateOrSetShoppingList(key string, list *crdt_go.ShoppingList) bool {
	readList, listExists := db.getShoppingList(key)

	if listExists {
		// merge and store
		readList.Merge(list)

		crdtBytes, err := json.Marshal(readList)

		if err != nil {
			return false
		}

		return db.storeValue([]byte(key), crdtBytes)
	} else {
		// simply store

		crdtBytes, err := json.Marshal(list)

		if err != nil {
			return false
		}

		return db.storeValue([]byte(key), crdtBytes)
	}

}

/**
* Gets a shopping list from the database
 */
func (db *DatabaseInstance) getShoppingList(key string) (*crdt_go.ShoppingList, bool) {
	crdtBytes, readSuccess := db.getValue([]byte(key))

	if !readSuccess {
		return nil, false
	}

	var crdt *crdt_go.ShoppingList

	err := json.Unmarshal(crdtBytes, &crdt)

	if err != nil {
		return nil, false
	}

	return crdt, true
}

/**
* Gets a value from the database
 */
func (db *DatabaseInstance) storeValue(key []byte, value []byte) bool {
	log.Println("wrote list", string(key))

	db.lock.Lock()

	err := db.conn.Store(key, value)

	if err != nil {
		db.lock.Unlock()
		return false
	}

	if db.conn.Commit() != nil {
		db.lock.Unlock()
		return false
	}

	db.lock.Unlock()

	return true
}

/**
* Stores a value into the database
 */
func (db *DatabaseInstance) getValue(key []byte) ([]byte, bool) {
	log.Println("read list", string(key))

	data, err := db.conn.Fetch(key)

	if err != nil {
		return []byte{}, false
	}

	if db.conn.Commit() != nil {
		return []byte{}, false
	}

	return data, true
}
