package main

import "sync"

/**
* Until unqlite is implemented, this will only be for debugging purposes
 */

type DatabaseInstance struct {
	data map[string][]byte
	lock sync.Mutex
}

func (db *DatabaseInstance) initialize() {
	db.data = make(map[string][]byte)
}

func (db *DatabaseInstance) writeToKey(key string, data []byte) {
	db.lock.Lock()

	db.data[key] = data

	db.lock.Unlock()
}

/**
* Returns the value of the specified key in bytes
 */
func (db *DatabaseInstance) getValueRaw(key string) []byte {
	return db.data[key]
}
