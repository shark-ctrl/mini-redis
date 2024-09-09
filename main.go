package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

var server redisServer
var wg sync.WaitGroup

func main() {

	wg.Add(1)

	//initialize redis server and configuration
	loadServerConfig()
	initServerConfig()
	initServer()

	//listen to the shutdown signal
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func(server *redisServer) {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeRedisServer()
			//modifying atomic variables means that the server is ready to shut down.
			server.done.Store(1)
		}
	}(&server)

	//parse address information
	address := server.ip + ":" + strconv.Itoa(server.port)
	log.Println("this redis server address:", address)
	//binding port number
	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("redis server listen failed,err:", err)
		return
	}
	server.listen = listen

	//handle client connections that are actively closed
	go func(server *redisServer) {
		for {
			c := <-server.closeClientCh
			log.Println("receive close client signal")
			_ = c.conn.Close()
			server.clients.Delete(c.string())
			log.Println("close client successful ", server.clients)
		}
	}(&server)

	go func(s *redisServer) {
		for redisClient := range s.commandCh {
			processCommand(&redisClient)
		}
	}(&server)

	//listen for incoming connections.
	go func() {
		for {
			log.Println("event loop is listening and waiting for client connection.")
			conn, err := listen.Accept()
			if err != nil {
				log.Println("accept conn failed,err:", err)
				break
			}
			acceptTcpHandler(conn)

		}
	}()

	wg.Wait()
	log.Println("redis service is down........................")

}
