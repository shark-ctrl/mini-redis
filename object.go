package main

import (
	"strconv"
)

const (
	/* Error codes */
	REDIS_OK  = 0
	REDIS_ERR = -1
)

func createStringObject(ptr *string, len int) *robj {
	return createEmbeddedStringObject(ptr, len)
}

func createEmbeddedStringObject(ptr *string, len int) *robj {
	o := new(robj)
	o.robjType = REDIS_STRING
	o.encoding = REDIS_ENCODING_EMBSTR
	i := interface{}(*ptr)
	o.ptr = &i
	return o
}

func tryObjectEncoding(o *robj) *robj {
	var value int64
	var s string
	var sLen int

	s = (*o.ptr).(string)
	sLen = len(s)

	if sLen < 21 && string2l(&s, sLen, &value) {
		if value >= 0 && value < REDIS_SHARED_INTEGERS {
			return shared.integers[value]
		} else {
			o.encoding = REDIS_ENCODING_INT
			num := interface{}(value)
			o.ptr = &num
		}
	}

	return o
}

func createObject(oType int, ptr *interface{}) *robj {
	o := new(robj)
	o.robjType = oType
	o.encoding = REDIS_ENCODING_RAW
	o.ptr = ptr
	return o
}

func createListObject() *robj {
	l := listCreate()
	i := interface{}(l)
	o := createObject(REDIS_LIST, &i)
	o.encoding = REDIS_ENCODING_LINKEDLIST
	return o
}

func getLongFromObjectOrReply(c *redisClient, o *robj, target *int64, msg *string) bool {
	value, err := strconv.ParseInt((*o.ptr).(string), 10, 64)
	if err != nil {
		if msg != nil {
			addReplyError(c, *msg)
		} else {
			addReplyError(c, "value is not an integer or out of range")
		}
		return false
	}
	*target = value
	return true
}

func checkType(c *redisClient, o *robj, rType int) bool {
	if o.robjType != rType {
		addReply(c, shared.wrongtypeerr)
		return false
	}
	return true
}

func getLongLongFromObjectOrReply(c *redisClient, expire string, target *int64, msg string) int {
	var value int64
	value, err := strconv.ParseInt(expire, 10, 64)
	if err != nil {
		addReply(c, shared.err)
		return REDIS_ERR
	}
	if value < 0 {
		addReplyError(c, "value is not an integer or out of range")
		return REDIS_ERR
	}
	*target = value
	return REDIS_OK
}
