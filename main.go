package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
)

type RedisServer struct {
	ip            string
	port          int
	shutDownCh    chan struct{}
	closeClientCh chan RedisClient
	done          atomic.Int32
	clients       sync.Map
	listen        net.Listener
}

func (server *RedisServer) initServer() {
	log.Println("init redis server")
	server.ip = "localhost"
	server.port = 6379
	server.shutDownCh = make(chan struct{})
	server.closeClientCh = make(chan RedisClient)
}

func (server *RedisServer) loadServerConfig() {
	log.Println("load redis server config")
}

func (server *RedisServer) acceptTcpHandler(conn net.Conn, ch chan RedisClient) {

	if server.done.Load() == 1 {
		log.Println("the current service is being shut down. The connection is denied.")
		_ = conn.Close()

	}

	c := &RedisClient{Conn: conn}
	server.clients.Store(c, struct{}{})

	go c.ReadQueryFromClient(ch)

}

func (server *RedisServer) close() {
	log.Println("close listen and all redis client")
	_ = server.listen.Close()
	server.clients.Range(func(key, value any) bool {
		c := key.(*RedisClient)
		_ = c.Conn.Close()
		server.clients.Delete(c)
		return true
	})
}

type RedisClient struct {
	Conn net.Conn
}

func (c *RedisClient) ReadQueryFromClient(ch chan RedisClient) {
	reader := bufio.NewReader(c.Conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil && err == io.EOF {
			log.Println("the redis client has been closed")
			ch <- *c
			break
		} else if err != nil {
			log.Println("redis client receive msg error :", err)
		}

		log.Println("receive msg:", msg)
		if msg == "ping\r\n" {
			pong := []byte("pong\r\n")
			_, _ = c.Conn.Write(pong)

		}

	}
}

func main() {
	//创建redis服务器
	server := new(RedisServer)
	//服务和配置初始化
	server.initServer()
	server.loadServerConfig()
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
			log.Println("server info ", server)
		}
	}(server)
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
		server.close()
		log.Println("redis server shutdown successful,server :", server)
	}(server)

	//监听关闭的客户端
	go func(server *RedisServer) {
		for {
			c := <-server.closeClientCh
			log.Println("receive close client signal")
			_ = c.Conn.Close()
			server.clients.Delete(c)
			log.Println("close client successful ", server.clients)
		}
	}(server)

	//阻塞监听处理连接
	for {

		log.Println("event loop is listening and waiting for client connection.")
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal("accept conn failed,err:", err)
			continue
		}
		server.acceptTcpHandler(conn, server.closeClientCh)

	}

}
