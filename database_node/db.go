package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"crypto/sha256"

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
	var list_store_res bool 
	var list_store_context_res bool 
	specialKey := []byte("lists_id_dot_contents")
	currentContentsBytes, readSuccess := db.getValue(specialKey)

	if listExists {
		// merge and store
		readList.Merge(list)

		crdtBytes, err := json.Marshal(readList)

		if err != nil {
			return false
		}

		list_store_res:= db.storeValue([]byte(key), crdtBytes)
		dot_context_hash, err := hashOfAWSetContext(readList.AwSet)

		if err != nil {
			log.Printf("Error computing hash of AWSet context: %s", err)
			return false
		}

		// Update the lists_id_dot_contents record with the context hash
		list_store_context_res := db.updateOrSetListsIdDotContents(key, dot_context_hash)
		if !updateSuccess {
			log.Printf("Error updating lists_id_dot_contents for key %s", key)
		}


	} else {
		// simply store

		crdtBytes, err := json.Marshal(list)

		if err != nil {
			return false
		}

		list_store_res:= db.storeValue([]byte(key), crdtBytes)
		dot_context_hash, err := hashOfAWSetContext(list.AwSet)

		if err != nil {
			log.Printf("Error computing hash of AWSet context: %s", err)
			return false
		}

		list_store_context_res := db.updateOrSetListsIdDotContents(key, dot_context_hash)
		if !updateSuccess {
			log.Printf("Error updating lists_id_dot_contents for key %s", key)
		}
	}

	return (storeSuccess && list_store_context_res)

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


//Logic for anti-entropy
func (db *DatabaseInstance) updateOrSetListsIdDotContents(key string, contextHash string) bool {
    specialKey := []byte("lists_id_dot_contents")

    
    currentContentsBytes, readSuccess := db.getValue(specialKey)
    var currentContents map[string]string
    if readSuccess {
        if err := json.Unmarshal(currentContentsBytes, &currentContents); err != nil {
            log.Printf("Error unmarshaling current lists_id_dot_contents: %s", err)
            return false
        }
    } else {
        // Initialize if the record does not exist
        currentContents = make(map[string]string)
    }

    // Update or set the entry for the current shopping list with the context hash
    currentContents[key] = contextHash

    
    updatedContentsBytes, err := json.Marshal(currentContents)
    if err != nil {
        log.Printf("Error marshaling updated lists_id_dot_contents: %s", err)
        return false
    }

    return db.storeValue(specialKey, updatedContentsBytes)
}

func (db *DatabaseInstance) GetAllListsIdDotContents() (map[string]string, error) {
    specialKey := []byte("lists_id_dot_contents")
    
    
    contentsBytes, readSuccess := db.getValue(specialKey)
    if !readSuccess {
        
        return make(map[string]string), nil
    }

    // Unmarshal the JSON data into a map
    var listsIdDotContents map[string]string
    err := json.Unmarshal(contentsBytes, &listsIdDotContents)
    if err != nil {
        log.Printf("Error unmarshaling lists_id_dot_contents: %s", err)
        return nil, err
    }

    return listsIdDotContents, nil
}


func hashOfAWSetContext(awset *crdt_go.AWSet) (string, error) {
    jsonData, err := json.Marshal(awset.Context)
    if err != nil {
        return "", err
    }

    hash := sha256.Sum256(jsonData)
    hexHash := fmt.Sprintf("%x", hash)
    return hexHash, nil
}