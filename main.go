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

var server RedisServer
var wg sync.WaitGroup

func main() {

	wg.Add(1)

	//服务和配置初始化
	initServer()
	loadServerConfig()
	initServerConfig()

	//监听系统关闭信号
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func(server *RedisServer) {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Println("receive close channel")
			server.shutDownCh <- struct{}{}
			server.done.Store(1)
		}
	}(&server)

	//解析地址信息
	address := server.ip + ":" + strconv.Itoa(server.port)
	log.Println("this redis server address:", address)
	//绑定端口
	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("redis server listen failed,err:", err)
		return
	}
	server.listen = listen

	//监听程序关闭
	go func(server *RedisServer) {
		<-server.shutDownCh
		log.Println("preparing to shut down the Redis server.")
		closeRedisService()
		log.Println("redis server shutdown successful,server :", server)
	}(&server)

	//监听客户端主动关闭
	go func(server *RedisServer) {
		for {
			c := <-server.closeClientCh
			log.Println("receive close client signal")
			_ = c.conn.Close()
			server.clients.Delete(c.string())
			log.Println("close client successful ", server.clients)
		}
	}(&server)

	go func(s *RedisServer) {
		for redisClient := range s.commandCh {
			processCommand(&redisClient)
		}
	}(&server)

	//阻塞监听处理连接
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
	log.Println("shutdown the redis service.......................")

}
