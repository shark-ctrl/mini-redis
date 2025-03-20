package main

import (
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

type redisCommandProc func(redisClient *redisClient)

type redisCommand struct {
	name  string
	proc  redisCommandProc
	arity int64
	sflag string
	flag  int
}

var redisCommandTable = []redisCommand{
	{name: "COMMAND", proc: commandCommand, arity: 0, sflag: "rlt", flag: 0},
	{name: "PING", proc: pingCommand, arity: 0, sflag: "rtF", flag: 0},
	{name: "SET", proc: setCommand, sflag: "rtF", flag: 0},
	{name: "GET", proc: getCommand, sflag: "rtF", flag: 0},
	{name: "RPUSH", proc: rpushCommand, sflag: "wmF", flag: 0},
	{name: "LRANGE", proc: lrangeCommand, sflag: "r", flag: 0},
	{name: "LINDEX", proc: lindexCommand, sflag: "r", flag: 0},
	{name: "LPOP", proc: lpopCommand, sflag: "wF", flag: 0},
	{name: "HSET", proc: hsetCommand, arity: 4, sflag: "wmF", flag: 0},
	{name: "HMSET", proc: hmsetCommand, arity: -4, sflag: "wm", flag: 0},
	{name: "HSETNX", proc: hsetnxCommand, arity: 4, sflag: "wm", flag: 0},
	{name: "HGET", proc: hgetCommand, arity: 3, sflag: "rF", flag: 0},
	{name: "HMGET", proc: hmgetCommand, arity: -3, sflag: "r", flag: 0},
	{name: "HGETALL", proc: hgetallCommand, arity: 2, sflag: "r", flag: 0},
	{name: "HDEL", proc: hdelCommand, arity: -3, sflag: "wF", flag: 0},
	{name: "ZADD", proc: zaddCommand, arity: -4, sflag: "wmF", flag: 0},
	{name: "ZREM", proc: zremCommand, arity: -3, sflag: "wF", flag: 0},
	{name: "ZCARD", proc: zcardCommand, arity: 2, sflag: "rF", flag: 0},
	{name: "ZRANK", proc: zrankCommand, arity: 3, sflag: "rF", flag: 0},
	{name: "INCR", proc: incrCommand, arity: 2, sflag: "wmF", flag: 0},
}
var shared sharedObjectsStruct

type sharedObjectsStruct struct {
	crlf           *string
	ok             *string
	err            *string
	pong           *string
	syntaxerr      *string
	nullbulk       *string
	wrongtypeerr   *string
	czero          *string
	cone           *string
	colon          *string
	emptymultibulk *string
	integers       [REDIS_SHARED_INTEGERS]*robj
	bulkhdr        [REDIS_SHARED_BULKHDR_LEN]*robj
}

func commandCommand(c *redisClient) {
	reply := "*" + strconv.Itoa(len(server.commands)) + *shared.crlf
	for _, command := range server.commands {
		reply += "$" + strconv.Itoa(len(command.name)) + *shared.crlf + command.name + *shared.crlf
	}

	addReply(c, &reply)
}

func pingCommand(c *redisClient) {
	addReply(c, shared.pong)
}

func setCommand(c *redisClient) {
	var j uint64
	var expire string
	unit := UNIT_SECONDS
	flags := REDIS_SET_NO_FLAGS
	//traverse the arguments after setting the key-value pair in a set.
	for j = 3; j < c.argc; j++ {
		ptr := c.argv[j].ptr
		a := (*ptr).(string)
		var next string
		if j == c.argc-1 {
			next = ""
		} else {
			nextPtr := c.argv[j+1].ptr
			next = (*nextPtr).(string)
		}
		// if the string "nx" is included, mark through bitwise operations that the current key can only be set when it does not exist.
		if strings.ToLower(a) == "nx" {
			flags |= REDIS_SET_NX
		} else if strings.ToLower(a) == "xx" { //if "xx" is included, mark the flags to indicate that the key can only be set if it already exists.
			flags |= REDIS_SET_XX
		} else if strings.ToLower(a) == "ex" { //if it is "ex", set the unit to seconds and read the next parameter.
			unit = UNIT_SECONDS
			expire = next
			j++
		} else if strings.ToLower(a) == "px" { //if it is "px", set the unit to milliseconds and read the next parameter.
			unit = UNIT_MILLISECONDS
			expire = next
			j++
		} else { // Treat all other cases as exceptions.
			addReply(c, shared.syntaxerr)
			return
		}
	}
	//pass the key, value, instruction identifier flags, and expiration time unit into `setGenericCommand` for memory persistence operation.
	setGenericCommand(c, flags, c.argv[1], c.argv[2], expire, unit, "", "")
}

func setGenericCommand(c *redisClient, flags int, key *robj, val *robj, expire string, unit int, ok_reply string, abort_reply string) {
	//initialize a pointer to record the expiration time in milliseconds.
	var milliseconds *int64
	milliseconds = new(int64)
	//if `expire` is not empty, parse it as an int64 and store it in `milliseconds`.
	if expire != "" {
		if getLongLongFromObjectOrReply(c, expire, milliseconds, nil) != REDIS_OK {
			return
		}

		if unit == UNIT_SECONDS {
			*milliseconds = *milliseconds * 1000
		}
	}
	/**
	the following two cases will no longer undergo key-value persistence operations:
	   1. if the command contains "nx" and the data exists for this value.
	   2. if the command contains "xx" and the data for this value does not exist.
	*/
	if (flags&REDIS_SET_NX > 0 && lookupKeyWrite(c.db, key) != nil) ||
		(flags&REDIS_SET_XX > 0 && lookupKeyWrite(c.db, key) == nil) {
		addReply(c, shared.nullbulk)
		return
	}
	//if `expire` is not empty, add the converted value to the current time to obtain the expiration time. Then,
	//use the passed key as the key and the expiration time as the value to store in the `expires` dictionary.
	if expire != "" {
		c.db.expires[(*key.ptr).(string)] = time.Now().UnixMilli() + *milliseconds
	}
	//store the key-value pair in a dictionary.
	c.db.dict[(*key.ptr).(string)] = val

	addReply(c, shared.ok)
}

func incrCommand(c *redisClient) {
	incrDecrCommand(c, 1)
}

func incrDecrCommand(c *redisClient, incr int64) {
	var value int64
	var oldValue int64
	var newObj *robj
	o := lookupKeyWrite(c.db, c.argv[1])

	if o != nil && checkType(c, o, REDIS_STRING) {
		return
	}

	var s string
	if o == nil {
		s = ""
	} else {
		s = (*o.ptr).(string)
	}

	if getLongLongFromObjectOrReply(c, s, &value, nil) != REDIS_OK {
		return
	}

	oldValue = value

	if (incr < 0 && oldValue < 0 && incr < (math.MinInt64-oldValue)) ||
		(incr > 0 && oldValue > 0 && incr > (math.MaxInt64-oldValue)) {
		errReply := "increment or decrement would overflow"
		addReplyError(c, &errReply)
		return
	}

	value += incr

	if o != nil &&
		(value < 0 || value >= REDIS_SHARED_INTEGERS) &&
		(value > math.MinInt64 || value < math.MaxInt64) {
		newObj = o

		i := interface{}(value)
		o.ptr = &i
	} else if o != nil {
		newObj = createStringObjectFromLongLong(value)
		dbOverwrite(c.db, c.argv[1], newObj)
	} else {
		newObj = createStringObjectFromLongLong(value)
		dbAdd(c.db, c.argv[1], newObj)
	}
	reply := *shared.colon + strconv.FormatInt(value, 10) + *shared.crlf
	addReply(c, &reply)

}

func createSharedObjects() {
	crlf := "\r\n"
	ok := "+OK\r\n"
	err := "-ERR\r\n"
	pong := "+PONG\r\n"
	syntaxerr := "-ERR syntax error\r\n"
	nullbulk := "$-1\r\n"
	wrongtypeerr := "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
	czero := ":0\r\n"
	cone := ":1\r\n"
	colon := ":"
	emptymultibulk := "*0\r\n"

	shared = sharedObjectsStruct{
		crlf:           &crlf,
		ok:             &ok,
		err:            &err,
		pong:           &pong,
		syntaxerr:      &syntaxerr,
		nullbulk:       &nullbulk,
		wrongtypeerr:   &wrongtypeerr,
		czero:          &czero,
		cone:           &cone,
		colon:          &colon,
		emptymultibulk: &emptymultibulk,
	}

	var i int64
	for i = 0; i < REDIS_SHARED_INTEGERS; i++ {
		num := interface{}(i)
		shared.integers[i] = createObject(REDIS_STRING, &num)
		shared.integers[i].encoding = REDIS_ENCODING_INT
	}

	for i := 0; i < REDIS_SHARED_BULKHDR_LEN; i++ {
		s := "*" + strconv.Itoa(i) + "\r\n"
		intf := interface{}(s)
		shared.bulkhdr[i] = createObject(REDIS_STRING, &intf)
	}
}

func selectDb(c *redisClient, id int) {
	c.db = &server.db[id]
}

func getCommand(c *redisClient) {
	getGenericCommand(c)
}

func getGenericCommand(c *redisClient) int {
	//check if the key exists, and if it does not, return a null bulk response from the constant values.
	o := lookupKeyReadOrReply(c, c.argv[1], shared.nullbulk)
	if o == nil {
		return REDIS_OK
	}
	//return the value to the client if it exists.
	addReplyBulk(c, o)
	return REDIS_OK
}

func rpushCommand(c *redisClient) {
	/**
	pass in the REDIS_TAIL flag to indicate that
	the current element should be appended to the tail of the list.
	*/
	pushGenericCommand(c, REDIS_TAIL)
}

func pushGenericCommand(c *redisClient, where int) {
	//check if the corresponding key exists.
	o := lookupKeyWrite(c.db, c.argv[1])
	var lobj *robj
	//if the key exists, then determine if it is a list.
	//if it is not, then throw an error exception.
	if o != nil && o.encoding != REDIS_ENCODING_LINKEDLIST {
		addReply(c, shared.wrongtypeerr)
		return
	} else if o != nil { //if it exists and is a list, then retrieve the Redis object for the list.
		lobj = o
	}
	//foreach element starting from index 2.
	var j uint64
	for j = 2; j < c.argc; j++ {
		//call `tryObjectEncoding` to perform special processing on the elements.
		c.argv[j] = tryObjectEncoding(c.argv[j])

		/**
		If the list is empty, then initialize it, create it, and store it in the Redis database.
		*/
		if lobj == nil {
			lobj = createListObject()
			dbAdd(c.db, c.argv[1], lobj)
		}
		/**
		pass in the list pointer, element pointer,
		and add flag to append the element to the head or tail of the list.
		*/
		listTypePush(lobj, c.argv[j], where)
	}
	//return the current length of the list.
	addReplyLongLong(c, (*lobj.ptr).(*list).len)
}

func lrangeCommand(c *redisClient) {
	var o *robj
	var start int64
	var end int64
	var llen int64
	var rangelen int64
	/**
	convert the strings at indexes 2 and 3 to numerical values.
	If an error occurs, respond with an exception and return.
	*/
	if !getLongFromObjectOrReply(c, c.argv[2], &start, nil) ||
		!getLongFromObjectOrReply(c, c.argv[3], &end, nil) {
		return
	}
	/**
	check if the linked list exists,
	and if it doesn't, respond with a null value.
	*/
	o = lookupKeyReadOrReply(c, c.argv[1], shared.emptymultibulk)

	/**
	check if the type is a linked list; if it is not, return a type error.
	*/
	if o == nil || checkType(c, o, REDIS_LIST) {
		return
	}
	//get the start and end values of a range query. If they are negative, add the length of the linked list.
	llen = (*o.ptr).(*list).len
	if start < 0 {
		start += llen
	}

	if start < 0 {
		start += llen
	}

	if end < 0 {
		end += llen
	}
	//if start is still less than 0, set it to 0.
	if start < 0 {
		start = 0
	}
	/**
	check if start is greater than the length of the linked list or if start is greater than end.
	If either of these exceptions occurs, respond with an error.
	*/
	if start >= llen || start > end {
		addReplyError(c, shared.emptymultibulk)
		return
	}
	//if end is greater than list length, set it to the list length.
	if end > llen {
		end = llen - 1
	}

	rangelen = end - start + 1
	addReplyMultiBulkLen(c, rangelen)

	if o.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
	} else if o.encoding == REDIS_ENCODING_LINKEDLIST {
		lobj := (*o.ptr).(*list)
		node := listIndex(lobj, start)
		//foreach the linked list starting from "start" based on "rangelen."
		for rangelen > 0 {
			rObj := (*node.value).(*robj)
			addReplyBulk(c, rObj)
			node = node.next
			rangelen--
		}

	} else {
		log.Panic("List encoding is not LINKEDLIST nor ZIPLIST!")
	}

}

func lindexCommand(c *redisClient) {
	/**
	check if the linked list exists; if it doesn't, return empty.
	*/
	o := lookupKeyReadOrReply(c, c.argv[1], shared.nullbulk)

	//verify if the type is a linked list.
	if o == nil || checkType(c, o, REDIS_LIST) {
		return
	}

	if o.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
	} else if o.encoding == REDIS_ENCODING_LINKEDLIST {
		/**
		retrieve the parameter at index 2 to obtain the index position,
		then fetch the element from the linked list at that index and return it.
		*/
		lobj := (*o.ptr).(*list)
		s := (*c.argv[2].ptr).(string)
		idx, _ := strconv.ParseInt(s, 10, 64)
		ln := listIndex(lobj, idx)

		if ln != nil {
			value := (*ln.value).(*robj)
			addReplyBulk(c, value)
		} else {
			addReply(c, shared.nullbulk)
		}
	} else {
		log.Panic("Unknown list encoding")
	}

}

func lpopCommand(c *redisClient) {
	// params is REDIS_HEAD, which means to retrieve the head element.
	popGenericCommand(c, REDIS_HEAD)
}

func popGenericCommand(c *redisClient, where int) {
	//check if the key exists, and if it doesn't, respond with an empty response.
	o := lookupKeyReadOrReply(c, c.argv[1], shared.nullbulk)

	//If the type is not a linked list, throw an exception and return.
	if o == nil || checkType(c, o, REDIS_LIST) {
		return
	}

	value := listTypePop(o, where)
	//retrieve the first element of the linked list based on the WHERE identifier.
	if value == nil {
		addReply(c, shared.nullbulk)
	} else {
		/**
		return the element value, and check if the linked list is empty.
		If it is empty, delete the key-value pair in the Redis database.
		*/
		addReplyBulk(c, value)
		if listTypeLength(o) == 0 {
			dbDelete(c.db, c.argv[1])
		}
	}
}

func hsetCommand(c *redisClient) {
	/**
	check if the dict object exists, and if it does not exist, create hash obj
	if it exists, then determine whether it is a hash obj. If not, return a type error.
	*/
	o := hashTypeLookupWriteOrCreate(c, c.argv[1])

	if o == nil {
		return
	}
	//try to convert strings that can be converted to numerical types into numerical types.
	hashTypeTryObjectEncoding(o, &c.argv[2], &c.argv[3])
	/**
	pass the information to the database to find the dictionary object along with the fields and values,
	return the dict update count
	*/
	update := hashTypeSet(o, c.argv[2], c.argv[3])
	//if it is an update operation, return 0; if it is the first insertion of a field, return 1.
	if update == 1 {
		addReply(c, shared.czero)
	} else {
		addReply(c, shared.cone)
	}
}

func hmsetCommand(c *redisClient) {
	/**
	determine if the  command params is singular
	if it is, respond with wrong number
	*/
	if c.argc%2 == 1 {
		errMsg := "wrong number of arguments for HMSET"
		addReplyError(c, &errMsg)
		return
	}
	/**
	starting from index 2,foreach key-value pair,
	perform encoding conversion, and save to dict obj
	*/
	var i uint64
	o := hashTypeLookupWriteOrCreate(c, c.argv[1])
	for i = 2; i < c.argc; i += 2 {
		hashTypeTryObjectEncoding(o, &c.argv[i], &c.argv[i+1])
		hashTypeSet(o, c.argv[i], c.argv[i+1])
	}

	addReply(c, shared.ok)
}

func hsetnxCommand(c *redisClient) {
	//perform dict lookup, type validation, and creation if it does not exist.
	o := hashTypeLookupWriteOrCreate(c, c.argv[1])
	//if it does not exist, return 0 and do not perform any operation.
	if hashTypeExists(o, c.argv[2]) {
		addReply(c, shared.czero)
		return
	}
	/**
		1. perform the field type and value type conversion.
	 	2. save field(argv[2])、value(argv[3]) to the dict obj
		3. respond to the client with the result 1
	*/
	hashTypeTryObjectEncoding(o, &c.argv[2], &c.argv[3])
	hashTypeSet(o, c.argv[2], c.argv[3])
	addReply(c, shared.cone)
}

func hgetCommand(c *redisClient) {
	//check if the dictionary exists, and if it does not exist, return null.
	o := lookupKeyReadOrReply(c, c.argv[1], shared.nullbulk)
	//if it is not a hash object, return a type error
	if o == nil || checkType(c, o, REDIS_HASH) {
		return
	}
	//if the corresponding field in the dictionary exists, return this value
	addHashFieldToReply(c, o, c.argv[2])

}

func addHashFieldToReply(c *redisClient, o *robj, field *robj) {
	//If the dictionary is empty, return nullbulk
	if o == nil {
		addReply(c, shared.nullbulk)
		return
	}

	if o.encoding == REDIS_ENCODING_ZIPLIST {
		//todo something
	} else if o.encoding == REDIS_ENCODING_HT {
		value := new(robj)
		/**
		pass the secondary pointer of the value to record the value corresponding to the field in the dictionary.
		if it is not null, return value; otherwise, return nullbulk.
		*/
		if hashTypeGetFromHashTable(o, field, &value) {
			addReplyBulk(c, value)
		} else {
			addReply(c, shared.nullbulk)
		}
	}

}

func hashTypeGetFromHashTable(o *robj, field *robj, value **robj) bool {
	dict := (*o.ptr).(map[string]*robj)
	key := (*field.ptr).(string)
	if v, e := dict[key]; e {
		*value = v
		return true

	}

	return false
}

func hmgetCommand(c *redisClient) {
	o := lookupKeyReadOrReply(c, c.argv[1], shared.nullbulk)
	if o == nil || checkType(c, o, REDIS_HASH) {
		return
	}

	addReplyMultiBulkLen(c, int64(c.argc-2))

	var i uint64
	//starting from the first field, read the fields from the dictionary and return them.
	for i = 2; i < c.argc; i++ {
		addHashFieldToReply(c, o, c.argv[i])
	}
}

func hgetallCommand(c *redisClient) {
	/**
	Use the XOR operation between the key and value to indicate that
	the current function needs to retrieve all keys and values from the dict.
	*/
	genericHgetallCommand(c, REDIS_HASH_KEY|REDIS_HASH_VALUE)

}

func genericHgetallCommand(c *redisClient, flags int) {
	multiplier := 0

	o := lookupKeyReadOrReply(c, c.argv[1], shared.emptymultibulk)
	if o == nil || checkType(c, o, REDIS_HASH) {
		return
	}
	/**
	if the result of the bitwise AND operation between the key and REDIS_HASH_KEY is greater than 0,
	increment the multiplier.
	*/
	if flags&REDIS_HASH_KEY > 0 {
		multiplier++
	}
	/**
	if the result of the bitwise AND operation between the key and REDIS_HASH_VALUE is greater than 0,
	increment the multiplier.
	*/
	if flags&REDIS_HASH_VALUE > 0 {
		multiplier++
	}
	/**
	response the client to return the value of the dictionary size multiplied by multiplier.

	*/
	dict := (*o.ptr).(map[string]*robj)
	l := len(dict)
	addReplyMultiBulkLen(c, int64(l*multiplier))
	//return the key-value pairs as required.
	for key, value := range dict {
		if flags&REDIS_HASH_KEY > 0 {
			i := interface{}(key)
			object := createObject(REDIS_STRING, &i)
			addReplyBulk(c, object)
		}

		if flags&REDIS_HASH_VALUE > 0 {
			addReplyBulk(c, value)
		}
	}

}

func hdelCommand(c *redisClient) {
	var deleted int64
	o := lookupKeyWriteOrReply(c, c.argv[1], shared.czero)
	if o == nil || checkType(c, o, REDIS_HASH) {
		return
	}
	//starting from index 2, locate all the keys.
	var i uint64
	for i = 2; i < c.argc; i++ {
		if hashTypeDelete(o, c.argv[i]) { //If the deletion is successful, increment the deleted counter.
			deleted++
		}
	}
	//If the dictionary has no key-value pairs after deletion, delete it directly.
	dict := (*o.ptr).(map[string]*robj)
	if len(dict) == 0 {
		dbDelete(c.db, c.argv[1])
	}

	addReplyLongLong(c, deleted)

}
