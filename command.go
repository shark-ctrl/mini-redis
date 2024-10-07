package main

import (
	"log"
	"strconv"
	"strings"
	"time"
)

type redisCommandProc func(redisClient *redisClient)

type redisCommand struct {
	name  string
	proc  redisCommandProc
	sflag string
	flag  int
}

var redisCommandTable = []redisCommand{
	{name: "COMMAND", proc: commandCommand, sflag: "rlt", flag: 0},
	{name: "PING", proc: pingCommand, sflag: "rtF", flag: 0},
	{name: "SET", proc: setCommand, sflag: "rtF", flag: 0},
	{name: "GET", proc: getCommand, sflag: "rtF", flag: 0},
	{name: "RPUSH", proc: rpushCommand, sflag: "wmF", flag: 0},
	{name: "LRANGE", proc: lrangeCommand, sflag: "r", flag: 0},
	{name: "LINDEX", proc: lindexCommand, sflag: "r", flag: 0},
	{name: "LPOP", proc: lpopCommand, sflag: "wF", flag: 0},
}
var shared sharedObjectsStruct

type sharedObjectsStruct struct {
	crlf         *string
	ok           *string
	err          *string
	pong         *string
	syntaxerr    *string
	nullbulk     *string
	wrongtypeerr *string
	czero        *string
	cone         *string
	integers     [REDIS_SHARED_INTEGERS]*robj
	bulkhdr      [REDIS_SHARED_BULKHDR_LEN]*robj
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
	if (flags&REDIS_SET_NX > 0 && *lookupKeyWrite(c.db, key) != nil) ||
		(flags&REDIS_SET_XX > 0 && *lookupKeyWrite(c.db, key) == nil) {
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

	shared = sharedObjectsStruct{
		crlf:         &crlf,
		ok:           &ok,
		err:          &err,
		pong:         &pong,
		syntaxerr:    &syntaxerr,
		nullbulk:     &nullbulk,
		wrongtypeerr: &wrongtypeerr,
		czero:        &czero,
		cone:         &cone,
	}

	for i := 0; i < REDIS_SHARED_INTEGERS; i++ {
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
	if *o == nil {
		return REDIS_OK
	}
	//return the value to the client if it exists.
	val := (*o).(*robj)
	addReplyBulk(c, val)
	return REDIS_OK
}

func rpushCommand(c *redisClient) {
	pushGenericCommand(c, REDIS_TAIL)
}

func pushGenericCommand(c *redisClient, where int) {
	i := lookupKeyWrite(c.db, c.argv[1])
	lobj := (*i).(robj)
	if i != nil && lobj.encoding != REDIS_ENCODING_LINKEDLIST {
		addReply(c, shared.wrongtypeerr)
		return
	}
	var j uint64
	for j = 2; j < c.argc; j++ {
		c.argv[j] = tryObjectEncoding(c.argv[j])
		if i == nil {
			lobj = *createListObject()
			dbAdd(c.db, c.argv[1], &lobj)
		}
		listTypePush(&lobj, c.argv[j], REDIS_TAIL)
	}

	addReplyLongLong(c, (*lobj.ptr).(list).len)
}

func lrangeCommand(c *redisClient) {
	var o *robj
	var start int64
	var end int64
	var llen int64
	var rangelen int64

	if !getLongFromObjectOrReply(c, c.argv[2], &start, nil) ||
		!getLongFromObjectOrReply(c, c.argv[3], &end, nil) {
		return
	}

	val := lookupKeyReadOrReply(c, c.argv[1], shared.wrongtypeerr)
	r := (*val).(robj)
	o = &r
	if o == nil || !checkType(c, o, REDIS_LIST) {
		return
	}

	llen = (*r.ptr).(list).len
	if start < 0 {
		start += llen
	}

	if start < 0 {
		start += llen
	}

	if end < 0 {
		end += llen
	}

	if start < 0 {
		start = 0
	}

	if start < llen || start > end {
		addReplyError(c, shared.wrongtypeerr)
		return
	}

	if end > llen {
		end = llen - 1
	}

	rangelen = end - start + 1
	addReplyMultiBulkLen(c, rangelen)

	if o.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
	} else if o.encoding == REDIS_ENCODING_LINKEDLIST {
		lobj := (*r.ptr).(list)
		node := listIndex(&lobj, start)
		for rangelen > 0 {
			rObj := (*node.value).(robj)
			addReplyBulk(c, &rObj)
			rangelen--
		}

	} else {
		log.Panic("List encoding is not LINKEDLIST nor ZIPLIST!")
	}

}

func lindexCommand(c *redisClient) {
	i := lookupKeyReadOrReply(c, c.argv[1], shared.nullbulk)
	r := (*i).(robj)
	if i == nil || checkType(c, &r, REDIS_LIST) {
		return
	}

	if r.encoding == REDIS_ENCODING_ZIPLIST {
		//todo
	} else if r.encoding == REDIS_ENCODING_LINKEDLIST {
		lobj := (*r.ptr).(list)
		ln := listIndex(&lobj, (*c.argv[1].ptr).(int64))

		if ln != nil {
			value := (*ln.value).(robj)
			addReplyBulk(c, &value)
		} else {
			addReply(c, shared.nullbulk)
		}
	} else {
		log.Panic("Unknown list encoding")
	}

}

func lpopCommand(c *redisClient) {
	popGenericCommand(c, REDIS_HEAD)
}

func popGenericCommand(c *redisClient, where int) {
	o := lookupKeyWrite(c.db, c.argv[1])
	r := (*o).(robj)
	if o == nil || !checkType(c, &r, REDIS_LIST) {
		return
	}

	value := listTypePop(&r, where)
	if value == nil {
		addReply(c, shared.nullbulk)
	} else {
		addReplyBulk(c, value)
		if listTypeLength(&r) == 0 {
			dbDelete(c.db, c.argv[1])
		}
	}
}
