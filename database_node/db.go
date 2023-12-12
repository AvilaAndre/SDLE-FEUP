package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
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
	log.Println("List/dotContext exist?", listExists," for key ", key)
	if listExists {
		//print read
		log.Println("Existed id ", readList)
		log.Println("Incoming id : ", list)
		// merge and store
		//Check first if the incoming list key is lists_id_dot_contents
		if key != "lists_id_dot_contents" {

			readList.Merge(list)

		}

		crdtBytes, err := json.Marshal(readList)

		if err != nil {
			//print error
			log.Println("Error marshalling list", err)
			return false
		}

		list_store_res := db.storeValue([]byte(key), crdtBytes)
		dot_context_hash, err := hashOfDotContext(readList.AwSet)
		if dot_context_hash == "" {
			fmt.Printf("Error computing hash of AWSet context: %v", err)
			return false
		}

		if err != nil {
			log.Printf("Error computing hash of AWSet context: %v", err)
			return false
		}
		// TODO: Check error here
		// Update the lists_id_dot_contents record with the context hash
		list_store_context_res := db.updateOrSetListsIdDotContents(key, dot_context_hash)
		if !list_store_context_res {
			log.Printf("Error updating lists_id_dot_contents for key %s", key)
		}
		return (list_store_res && list_store_context_res) //TODO: error can be here, if list_store_context_res is not true: check this

	} else {
		// simply store

		crdtBytes, err := json.Marshal(list)

		if err != nil {
			return false
		}
		list_store_res := db.storeValue([]byte(key), crdtBytes)

		dot_context_hash, err := hashOfDotContext(list.AwSet)
		if err != nil {
			log.Printf("Error computing hash of AWSet context: %s", err)
			return false
		}
		if list.AwSet == nil {
			return true // TODO: CHeck this
		}
		list_store_context_res := db.updateOrSetListsIdDotContents(key, dot_context_hash)
		if !list_store_context_res {
			log.Printf("Error updating lists_id_dot_contents for key %s", key)
		}
		return (list_store_res && list_store_context_res) //TODO: error can be here, if list_store_context_res is not true: check this
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
* Stores a value into the database
 */
func (db *DatabaseInstance) storeValue(key []byte, value []byte) bool {
	log.Println("wrote list or list_ids_dot_contents", string(key))

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
* Gets a value from the database
 */
func (db *DatabaseInstance) getValue(key []byte) ([]byte, bool) {
	log.Println("read somtething on getValue function in db.go", string(key))

	data, err := db.conn.Fetch(key)
	
	if err != nil {
		return []byte{}, false
	}
	
	if db.conn.Commit() != nil {
		return []byte{}, false
	}
	log.Println("This is the result of get Value", string(data))
	
	return data, true
}

/**
* Deletes a shopping list from the database
 */
func (db *DatabaseInstance) deleteList(key string) bool {
	
	success := db.deleteValue([]byte(key))
	

	return success
}

/**
* Deletes a key from the database
 */
func (db *DatabaseInstance) deleteValue(key []byte) bool {
	log.Println("delete list", string(key))
	
	if db.conn.Delete([]byte(key)) != nil {
		return false
	}

	if db.conn.Commit() != nil {
		return false
	}
	
	return true
}

// TODO: check if updateOrSetListsIdDotContents is used properly and works as expected
// Logic for anti-entropy
func (db *DatabaseInstance) updateOrSetListsIdDotContents(key string, contextHash string) bool {
	specialKey := []byte("lists_id_dot_contents")

	currentContentsBytes, readSuccess := db.getValue(specialKey)
	var currentContents map[string]string  = make(map[string]string)
	if !readSuccess {
		// Print This: If the lists_id_dot_contents record doesn't exist, create it to store list_id -> context_hash mappings
		currentContents[key] = contextHash
		fmt.Printf("If the lists_id_dot_contents record doesn't exist, create it to store list_id -> context_hash mappings, the_list_id_dot_context is: %s", currentContents)



	}else{
		err := json.Unmarshal(currentContentsBytes, &currentContents)
		if err != nil {
			log.Printf("Error unmarshaling lists_id_dot_contents: %s", err)
			return false
		} 
		//print the currentContents
		currentContents[key] = contextHash
		fmt.Printf("updateOrSetListsIdDotContents: the_list_id_dot_context is: %s", currentContents)
	
		
	
		
	}



	// Update or set the entry for the current shopping list with the context hash
	//Print the currentContents
	fmt.Printf("updateOrSetListsIdDotContents: the_list_id_dot_context is: %s", currentContents)
	updatedContentsBytes, err := json.Marshal(currentContents) //TODO: check if this is correct
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

//Usefull functions for future work

func hashOfDotContext(awset *crdt_go.AWSet) (string, error) {
	// Check if awset is nil
    if awset == nil {
        return "", errors.New("awset is nil")
    }

    // Check if awset.Context is nil
    if awset.Context == nil {
        return "", errors.New("awset.Context is nil")
    }

    // Serialize the dot_context item to JSON
    jsonData, err := json.Marshal(awset.Context)
    if err != nil {
        return "", err
    }

    // Compute the SHA-256 hash of the JSON string
    hash := sha256.Sum256(jsonData)

    // Convert the hash to a hexadecimal string
    hexHash := fmt.Sprintf("%x", hash)

    return hexHash, nil
}
