package main

type redisDb struct {
	dict    map[string]interface{}
	expires map[string]uint64
	id      int
}

func lookupKeyWrite(db *redisDb, key string) {

}

func expireIfNeeded(db *redisDb, key string) {

}

func lookupKey(db *redisDb, key string) {

}
