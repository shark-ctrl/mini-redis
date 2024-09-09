package main

import "strconv"

const (
	/* Error codes */
	REDIS_OK  = 0
	REDIS_ERR = -1
)

func getLongLongFromObjectOrReply(c *redisClient, expire string, target *uint64, msg string) int {
	var value uint64
	value, err := strconv.ParseUint(expire, 10, 64)
	if err != nil {
		addReply(c, "invalid value")
		return REDIS_ERR
	}
	if value < 0 {
		addReplyErrorLength(c, "value is not an integer or out of range")
		return REDIS_ERR
	}
	*target = value
	return REDIS_OK
}
