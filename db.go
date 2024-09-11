package main

import "time"

type redisDb struct {
	dict    map[string]interface{}
	expires map[string]int64
	id      int
}

func lookupKeyWrite(db *redisDb, key string) *interface{} {
	expireIfNeeded(db, key)
	return lookupKey(db, key)
}

func expireIfNeeded(db *redisDb, key string) int {
	when, exists := db.expires[key]
	if !exists {
		return 0
	}
	if when < 0 {
		return 0
	}
	now := time.Now().UnixMilli()

	if now < when {
		return 0
	}

	deDelete(db, key)

	return 1

}

func deDelete(db *redisDb, key string) {
	delete(db.expires, key)
	delete(db.dict, key)
}

func lookupKeyRead(db *redisDb, key string) *interface{} {
	expireIfNeeded(db, key)
	val := lookupKey(db, key)
	return val
}

func lookupKey(db *redisDb, key string) *interface{} {
	val := db.dict[key]
	return &val
}

func lookupKeyReadOrReply(c *redisClient, key string, reply *string) *interface{} {
	o := lookupKeyRead(c.db, key)
	if *o == nil {
		addReply(c, *reply)
	}
	return o
}
