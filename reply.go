package main

func addReply(c *redisClient, reply string) {
	c.conn.Write([]byte(reply))
}

func addReplyErrorLength(c *redisClient, s string) {
	c.conn.Write([]byte("-ERR\r\n" + s + "\r\n"))
}
