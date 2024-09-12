package main

import (
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
}
var shared sharedObjectsStruct

type sharedObjectsStruct struct {
	crlf      string
	ok        string
	err       string
	pong      string
	syntaxerr string
	nullbulk  string
}

func commandCommand(c *redisClient) {
	reply := "*" + strconv.Itoa(len(server.commands)) + shared.crlf
	for _, command := range server.commands {
		reply += "$" + strconv.Itoa(len(command.name)) + shared.crlf + command.name + shared.crlf
	}

	addReply(c, reply)
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
		a := c.argv[j]
		var next string
		if j == c.argc-1 {
			next = ""
		} else {
			next = c.argv[j+1]
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

func setGenericCommand(c *redisClient, flags int, key string, val string, expire string, unit int, ok_reply string, abort_reply string) {
	//initialize a pointer to record the expiration time in milliseconds.
	var milliseconds *int64
	milliseconds = new(int64)
	//if `expire` is not empty, parse it as an int64 and store it in `milliseconds`.
	if expire != "" {
		if getLongLongFromObjectOrReply(c, expire, milliseconds, "") != REDIS_OK {
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
		c.db.expires[key] = time.Now().UnixMilli() + *milliseconds
	}
	//store the key-value pair in a dictionary.
	c.db.dict[key] = val

	addReply(c, shared.ok)
}

func createSharedObjects() {
	shared = sharedObjectsStruct{
		crlf:      "\r\n",
		ok:        "+OK\r\n",
		err:       "-ERR\r\n",
		pong:      "+PONG\r\n",
		syntaxerr: "-ERR syntax error\r\n",
		nullbulk:  "$-1\r\n",
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
	o := lookupKeyReadOrReply(c, c.argv[1], &shared.nullbulk)
	if *o == nil {
		return REDIS_OK
	}
	//return the value to the client if it exists.
	val := (*o).(string)
	addReplyBulk(c, &val)
	return REDIS_OK
}
