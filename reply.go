package main

import "strconv"

func addReply(c *redisClient, reply string) {
	c.conn.Write([]byte(reply))
}

func addReplyBulk(c *redisClient, reply *string) {
	c.conn.Write([]byte("$" + strconv.Itoa(len(*reply)) + shared.crlf + *reply + shared.crlf))
}

func addReplyErrorLength(c *redisClient, s string) {
	c.conn.Write([]byte("-ERR\r\n" + s + "\r\n"))
}
