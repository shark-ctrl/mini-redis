package main

import "time"

type redisDb struct {
	dict    map[string]interface{}
	expires map[string]int64
	id      int
}

func lookupKeyWrite(db *redisDb, key string) *interface{} {
	//check if the key has expired, and if so, delete it.
	expireIfNeeded(db, key)
	//query the dictionary for the value corresponding to the key.
	return lookupKey(db, key)
}

func expireIfNeeded(db *redisDb, key string) int {
	//get the expiration time of the key.
	when, exists := db.expires[key]
	if !exists {
		return 0
	}
	if when < 0 {
		return 0
	}
	now := time.Now().UnixMilli()
	//if the current time is less than the expiration time, it means the current key has not expired, so return directly.
	if now < when {
		return 0
	}
	//delete expired keys.
	deDelete(db, key)

	return 1

}

func deDelete(db *redisDb, key string) {
	delete(db.expires, key)
	delete(db.dict, key)
}

func lookupKeyRead(db *redisDb, key string) *interface{} {
	//check if the key has expired and delete it.
	expireIfNeeded(db, key)
	val := lookupKey(db, key)
	return val
}

func lookupKey(db *redisDb, key string) *interface{} {
	val := db.dict[key]
	return &val
}

func lookupKeyReadOrReply(c *redisClient, key string, reply *string) *interface{} {
	//check if the key has expired to decide whether to delete the key from the dictionary, then query the dictionary for the result and return it.
	o := lookupKeyRead(c.db, key)
	if *o == nil {
		addReply(c, *reply)
	}
	return o
}
