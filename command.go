package main

import (
	"log"
	"strconv"
)

type redisCommandProc func(redisClient *redisClient)

type RedisCommand struct {
	name  string
	proc  redisCommandProc
	sflag string
	flag  int
}

var redisCommandTable = []RedisCommand{
	{name: "COMMAND", proc: CommandCommand, sflag: "rlt", flag: 0},
	{name: "PING", proc: PingCommand, sflag: "rtF", flag: 0},
}
var shared sharedObjectsStruct

type sharedObjectsStruct struct {
	crlf string
	ok   string
	err  string
	pong string
}

func CommandCommand(c *redisClient) {
	reply := "*" + strconv.Itoa(len(server.commands)) + shared.crlf
	for _, command := range server.commands {
		reply += "$" + strconv.Itoa(len(command.name)) + shared.crlf + command.name + shared.crlf
	}

	log.Println("command:" + reply)
	addReply(c, reply)
}

func PingCommand(c *redisClient) {
	addReply(c, shared.pong)
}

func createSharedObjects() {
	shared = sharedObjectsStruct{
		crlf: "\r\n",
		ok:   "+OK\r\n",
		err:  "-ERR\r\n",
		pong: "+PONG\r\n",
	}
}

func addReply(c *redisClient, reply string) {
	c.conn.Write([]byte(reply))
}
