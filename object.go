package main

import (
	"math"
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
	//get string value and length
	s = (*o.ptr).(string)
	sLen = len(s)
	/**
	If it can be converted into an integer and is between 0 and 10000,
	it is obtained from the constant pool.
	*/
	if sLen < 21 && string2l(&s, sLen, &value) {
		if value >= 0 && value < REDIS_SHARED_INTEGERS {
			return shared.integers[value]
		} else {
			/**
			If it is not within the scope of constant pool,
			it will be manually converted into an object of integer encoding type.
			*/
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
		errMsg := "value is not an integer or out of range"
		addReplyError(c, &errMsg)
		return false
	}

	if value < math.MinInt64 || value > math.MaxInt64 {
		if msg != nil {
			addReplyError(c, msg)
		} else {
			*msg = "value is not an integer or out of range"
			addReplyError(c, msg)
		}
		return false
	}
	*target = value
	return true
}

func checkType(c *redisClient, o *robj, rType int) bool {
	if o.robjType != rType {
		addReply(c, shared.wrongtypeerr)
		return true
	}
	return false
}

func getLongLongFromObjectOrReply(c *redisClient, expire string, target *int64, msg *string) int {
	var value int64
	value, err := strconv.ParseInt(expire, 10, 64)
	if err != nil {
		addReply(c, shared.err)
		return REDIS_ERR
	}
	if value < 0 {
		*msg = "value is not an integer or out of range"
		addReplyError(c, msg)
		return REDIS_ERR
	}
	*target = value
	return REDIS_OK
}
