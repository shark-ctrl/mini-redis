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

type RedisClient struct {
	conn         net.Conn
	argc         uint64
	argv         []string
	multibulklen int64
	reqType      int
	queryBuf     []byte
	cmd          RedisCommand
	lastCmd      RedisCommand
}

func (c *RedisClient) ReadQueryFromClient(CloseClientCh chan RedisClient, commandCh chan RedisClient) {
	reader := bufio.NewReader(c.conn)

	processInputBuffer(c, reader, CloseClientCh, commandCh)
}

func processInputBuffer(c *RedisClient, reader *bufio.Reader, CloseClientCh chan RedisClient, commandCh chan RedisClient) {
	for {
		c.multibulklen = -1
		bytes, err := reader.ReadBytes('\n')
		c.queryBuf = bytes
		if err != nil {
			log.Println("the redis client has been closed")
			CloseClientCh <- *c
			break
		}

		if len(c.queryBuf) == 0 || (len(c.queryBuf) >= 2 && c.queryBuf[len(c.queryBuf)-2] != '\r') {
			_, _ = c.conn.Write([]byte("-ERR unknown command\r\n"))
			log.Println("ERR unknown command")
			continue
		}

		if c.queryBuf[0] == '*' && c.multibulklen == -1 {
			c.reqType = REDIS_REQ_MULTIBULK
			c.multibulklen, err = strconv.ParseInt(string(c.queryBuf[1:len(c.queryBuf)-2]), 10, 32)
			if err != nil || c.multibulklen < 0 {
				_, _ = c.conn.Write([]byte("-ERR unknown command\r\n"))
				log.Println("ERR unknown command")
				continue
			}
			c.argv = make([]string, c.multibulklen)
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

func processMultibulkBuffer(c *RedisClient, reader *bufio.Reader, CloseClientCh chan RedisClient) error {
	c.argc = 0

	ll := int64(-1)

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

		if c.queryBuf[0] == '$' && ll < 0 {
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

			c.argv[c.argc] = string(c.queryBuf[0 : len(c.queryBuf)-2])
			c.argc++
		} else if c.queryBuf[0] == '$' && ll > 0 {
			return errors.New("ERR unknown command")
		} else if c.queryBuf[0] != '$' && ll < 0 { //未解析到长度就遇到其他的字符
			return errors.New("ERR unknown command")
		}
	}

	return nil
}

func (c RedisClient) string() string {
	return fmt.Sprintf("%#v", c)
}
