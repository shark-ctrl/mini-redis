package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

const (
	/* Client request types */
	REDIS_REQ_INLINE    = 1
	REDIS_REQ_MULTIBULK = 2
)

type redisClient struct {
	//redis client connection info
	conn         net.Conn
	argc         uint64
	argv         []string
	multibulklen int64
	reqType      int
	queryBuf     []byte
	cmd          redisCommand
	lastCmd      redisCommand
	db           *redisDb
}

func readQueryFromClient(c *redisClient, CloseClientCh chan redisClient, commandCh chan redisClient) {
	//get the network reader  through the redis client's connection.
	reader := bufio.NewReader(c.conn)
	//parse the string through the reader, and pass the parsing result to commandCh for Redis server to parse and execute.
	processInputBuffer(c, reader, CloseClientCh, commandCh)
}

func processInputBuffer(c *redisClient, reader *bufio.Reader, CloseClientCh chan redisClient, commandCh chan redisClient) {
	for {
		//initialize the array length to -1.
		c.multibulklen = -1
		//split each string by '\n'
		bytes, err := reader.ReadBytes('\n')
		c.queryBuf = bytes
		if err != nil {
			log.Println("the redis client has been closed")
			CloseClientCh <- *c
			break
		}
		//throw an exception if '\n' is not preceded by '\r'.
		if len(c.queryBuf) == 0 || (len(c.queryBuf) >= 2 && c.queryBuf[len(c.queryBuf)-2] != '\r') {
			_, _ = c.conn.Write([]byte("-ERR unknown command\r\n"))
			log.Println("ERR unknown command")
			continue
		}
		//If it starts with "*", it indicates a multiline string.
		if c.queryBuf[0] == '*' && c.multibulklen == -1 {
			//set the request type to multiline
			c.reqType = REDIS_REQ_MULTIBULK
			//get the length of the array based on the number following '*'
			c.multibulklen, err = strconv.ParseInt(string(c.queryBuf[1:len(c.queryBuf)-2]), 10, 32)
			if err != nil || c.multibulklen < 0 {
				_, _ = c.conn.Write([]byte("-ERR unknown command\r\n"))
				log.Println("ERR unknown command")
				continue
			}
			//based on the parsed length, initialize the size of the array.
			c.argv = make([]string, c.multibulklen)
			//based on the length indicated by "*", start parsing the string.
			e := processMultibulkBuffer(c, reader, CloseClientCh)
			if e != nil {
				_, _ = c.conn.Write([]byte("-ERR unknown command\r\n"))
				log.Println("ERR unknown command")
				continue
			} else {
				commandCh <- *c
			}
		} else if c.queryBuf[0] == '*' && c.multibulklen > -1 {
			_, _ = c.conn.Write([]byte("-ERR unknown command\r\n"))
			log.Println("ERR unknown command")
			continue
		} else {
			//todo the processing logic for single-line instructions is to be completed subsequently
			c.multibulklen = REDIS_REQ_INLINE
			_, _ = c.conn.Write([]byte("-ERR unknown command\r\n"))
			continue
		}

	}

}

func processMultibulkBuffer(c *redisClient, reader *bufio.Reader, CloseClientCh chan redisClient) error {
	c.argc = 0
	//initialize "ll" to record the length following each "$", then fetch the string based on this length.
	ll := int64(-1)
	//perform a for loop based on "multibulklen".
	for i := 0; i < int(c.multibulklen); i++ {
		bytes, e := reader.ReadBytes('\n')
		c.queryBuf = bytes
		if e != nil && e == io.EOF {
			log.Println("the redis client has been closed")
			CloseClientCh <- *c
			break
		} else if e != nil {
			return e
		}

		if len(c.queryBuf) == 0 || !(len(c.queryBuf)-2 >= 0 && c.queryBuf[len(c.queryBuf)-2] == '\r') {
			return errors.New("ERR unknown command")
		}
		//if a "$" is intercepted in this line, store the following numerical value in "ll".
		if c.queryBuf[0] == '$' {
			ll, e = strconv.ParseInt(string(c.queryBuf[1:len(c.queryBuf)-2]), 10, 32)
			if e != nil || ll <= 0 {
				return e
			}
			strBytes, e := reader.ReadBytes('\n')
			c.queryBuf = strBytes

			if e != nil {
				return e
			}

			if len(c.queryBuf) == 0 || int64(len(c.queryBuf))-2 != ll {
				return errors.New("ERR unknown command")
			}
			//parse and extract a string of specified length based on the value of "ll", store it in "argv", and then increment "argc".
			c.argv[c.argc] = string(c.queryBuf[0 : len(c.queryBuf)-2])
			c.argc++
		} else if c.queryBuf[0] != '$' && ll < 0 { //invalid str
			return errors.New("ERR unknown command")
		}
	}

	return nil
}

func (c redisClient) string() string {
	return fmt.Sprintf("%#v", c)
}
