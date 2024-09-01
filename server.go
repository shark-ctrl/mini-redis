package main

import (
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	REDIS_CMD_WRITE           = 1    /* "w" flag */
	REDIS_CMD_READONLY        = 2    /* "r" flag */
	REDIS_CMD_DENYOOM         = 4    /* "m" flag */
	REDIS_CMD_NOT_USED_1      = 8    /* no longer used flag */
	REDIS_CMD_ADMIN           = 16   /* "a" flag */
	REDIS_CMD_PUBSUB          = 32   /* "p" flag */
	REDIS_CMD_NOSCRIPT        = 64   /* "s" flag */
	REDIS_CMD_RANDOM          = 128  /* "R" flag */
	REDIS_CMD_SORT_FOR_SCRIPT = 256  /* "S" flag */
	REDIS_CMD_LOADING         = 512  /* "l" flag */
	REDIS_CMD_STALE           = 1024 /* "t" flag */
	REDIS_CMD_SKIP_MONITOR    = 2048 /* "M" flag */
	REDIS_CMD_ASKING          = 4096 /* "k" flag */
	REDIS_CMD_FAST            = 8192 /* "F" flag */
	/* Command call flags, see call() function */
	REDIS_CALL_NONE      = 0
	REDIS_CALL_SLOWLOG   = 1
	REDIS_CALL_STATS     = 2
	REDIS_CALL_PROPAGATE = 4
	REDIS_CALL_FULL      = (REDIS_CALL_SLOWLOG | REDIS_CALL_STATS | REDIS_CALL_PROPAGATE)
)

type redisServer struct {
	//record the ip and port number of the redis server.
	ip   string
	port int
	//semaphore used to notify shutdown.
	shutDownCh    chan struct{}
	commandCh     chan redisClient
	closeClientCh chan redisClient
	done          atomic.Int32
	//record all connected clients.
	clients sync.Map
	//listen and process new connections.
	listen   net.Listener
	commands map[string]RedisCommand
}

func initServer() {
	log.Println("init redis server")
	server.ip = "localhost"
	server.port = 6379
	server.shutDownCh = make(chan struct{})
	server.closeClientCh = make(chan redisClient)
	server.commandCh = make(chan redisClient)
	server.commands = make(map[string]RedisCommand)

	createSharedObjects()
}

func loadServerConfig() {
	log.Println("load redis server config")
}

func acceptTcpHandler(conn net.Conn) {
	//the current server is being or has been shut down, and no new connections are being processed.
	if server.done.Load() == 1 {
		log.Println("the current service is being shut down. The connection is denied.")
		_ = conn.Close()

	}
	//init the redis client and handles network read and write events.
	c := &redisClient{conn: conn, argc: 0, argv: make([]string, 0), multibulklen: -1}
	server.clients.Store(c.string(), c)
	go readQueryFromClient(c, server.closeClientCh, server.commandCh)

}

func closeRedisServer() {
	log.Println("close listen and all redis client")
	_ = server.listen.Close()
	server.clients.Range(func(key, value any) bool {
		client := value.(*redisClient)
		_ = client.conn.Close()
		server.clients.Delete(key)
		return true
	})

	wg.Done()
}

func initServerConfig() {
	populateCommandTable()
}

func populateCommandTable() {
	for i := 0; i < len(redisCommandTable); i++ {
		redisCommand := redisCommandTable[i]
		for _, f := range redisCommand.sflag {
			if f == 'w' {
				redisCommand.flag |= REDIS_CMD_WRITE
			} else if f == 'r' {
				redisCommand.flag |= REDIS_CMD_READONLY
			} else if f == 'm' {
				redisCommand.flag |= REDIS_CMD_DENYOOM
			} else if f == 'a' {
				redisCommand.flag |= REDIS_CMD_ADMIN
			} else if f == 'p' {
				redisCommand.flag |= REDIS_CMD_PUBSUB
			} else if f == 's' {
				redisCommand.flag |= REDIS_CMD_NOSCRIPT
			} else if f == 'R' {
				redisCommand.flag |= REDIS_CMD_RANDOM
			} else if f == 'S' {
				redisCommand.flag |= REDIS_CMD_SORT_FOR_SCRIPT
			} else if f == 'l' {
				redisCommand.flag |= REDIS_CMD_LOADING
			} else if f == 't' {
				redisCommand.flag |= REDIS_CMD_STALE
			} else if f == 'M' {
				redisCommand.flag |= REDIS_CMD_SKIP_MONITOR
			} else if f == 'K' {
				redisCommand.flag |= REDIS_CMD_ASKING
			} else if f == 'F' {
				redisCommand.flag |= REDIS_CMD_FAST
			} else {
				log.Panicln("Unsupported command flag")
			}

			server.commands[redisCommand.name] = redisCommand
		}

	}
}

func processCommand(c *redisClient) {
	redisCommand, exists := server.commands[strings.ToUpper(c.argv[0])]
	if !exists {
		c.conn.Write([]byte("-ERR unknown command\r\n"))
		return
	}
	c.cmd = redisCommand
	c.lastCmd = redisCommand

	call(c, REDIS_CALL_FULL)
}

func call(c *redisClient, flags int) {
	c.cmd.proc(c)

	//todo aof use flags
}
