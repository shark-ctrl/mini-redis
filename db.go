package main

import "time"

type redisDb struct {
	dict    map[string]*robj
	expires map[string]int64
	id      int
}

func lookupKeyWriteOrReply(c *redisClient, key *robj, reply *string) *robj {
	o := lookupKeyWrite(c.db, key)
	if o == nil {
		addReply(c, reply)
	}

	return o
}

func lookupKeyWrite(db *redisDb, key *robj) *robj {
	//check if the key has expired, and if so, delete it.
	expireIfNeeded(db, key)
	//query the dictionary for the value corresponding to the key.
	return lookupKey(db, key)
}

func expireIfNeeded(db *redisDb, key *robj) int {
	//get the expiration time of the key.
	when, exists := db.expires[(*key.ptr).(string)]
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

func deDelete(db *redisDb, key *robj) {
	delete(db.expires, (*key.ptr).(string))
	delete(db.dict, (*key.ptr).(string))
}

func lookupKeyRead(db *redisDb, key *robj) *robj {
	//check if the key has expired and delete it.
	expireIfNeeded(db, key)
	val := lookupKey(db, key)
	return val
}

func lookupKey(db *redisDb, key *robj) *robj {
	val := db.dict[(*key.ptr).(string)]
	return val
}

func lookupKeyReadOrReply(c *redisClient, key *robj, reply *string) *robj {
	//check if the key has expired to decide whether to delete the key from the dictionary, then query the dictionary for the result and return it.
	o := lookupKeyRead(c.db, key)
	if o == nil {
		addReply(c, reply)
	}
	return o
}

func dbAdd(db *redisDb, key *robj, val *robj) {
	db.dict[(*key.ptr).(string)] = val
}

func dbDelete(db *redisDb, key *robj) int64 {
	delete(db.dict, (*key.ptr).(string))
	return 1
}
