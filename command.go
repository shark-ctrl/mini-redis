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

	log.Println("command:" + reply)
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

	for j = 3; j < c.argc; j++ {
		a := c.argv[j]
		var next string
		if j == c.argc-1 {
			next = ""
		} else {
			next = c.argv[j+1]
		}

		if strings.ToLower(a) == "nx" {
			flags |= REDIS_SET_NX
		} else if strings.ToLower(a) == "xx" {
			flags |= REDIS_SET_XX
		} else if strings.ToLower(a) == "ex" {
			unit = UNIT_SECONDS
			expire = next
			j++
		} else if strings.ToLower(a) == "px" {
			unit = UNIT_MILLISECONDS
			expire = next
			j++
		} else {
			addReply(c, shared.syntaxerr)
			return
		}
	}

	setGenericCommand(c, flags, c.argv[1], c.argv[2], expire, unit, "", "")
}

func setGenericCommand(c *redisClient, flags int, key string, val string, expire string, unit int, ok_reply string, abort_reply string) {
	var milliseconds *int64
	if expire != "" {
		if getLongLongFromObjectOrReply(c, expire, milliseconds, "") != REDIS_OK {
			return
		}

		if unit == UNIT_SECONDS {
			*milliseconds = *milliseconds * 1000
		}
	}

	if (flags&REDIS_SET_NX > 0 && lookupKeyWrite(c.db, key) != nil) ||
		(flags&REDIS_SET_XX > 0 && lookupKeyWrite(c.db, key) == nil) {
		addReply(c, shared.nullbulk)
		return
	}

	if expire != "" {
		c.db.dict[key] = time.Now().UnixMilli() + *milliseconds
	}
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

func getGenericCommand(c *redisClient) {
	i, e := c.db.dict[c.argv[1]]
	if !e {
		addReply(c, shared.nullbulk)
		return
	}
	addReply(c, "$"+strconv.Itoa(len(i.(string)))+shared.crlf+i.(string)+shared.crlf)
}
