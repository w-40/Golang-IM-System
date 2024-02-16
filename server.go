package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex // 全局map，要加锁，sync是go的同步包

	//消息广播的channel
	Message chan string
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给所有的在线User
func (server *Server) ListenMessager() {
	for {
		msg := <-server.Message

		//将msg发送给全部的在线User
		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

func (server *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	//fmt.Println("链接建立成功")

	user := NewUser(conn, server)

	user.Online()

	//监听用户是否活跃的channel
	isAlive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read error:", err)
				return
			}
			// 提取用户的消息(去除'\n')
			msg := string(buf[:n-1])

			// 用户针对msg进行消息处理
			user.DoMessage(msg)

			//用户的任意消息，代表当前用户是活跃的
			isAlive <- true
		}
	}()

	// 当前Handler阻塞
	for {
		select {
		case <-isAlive:
			//当前用户是活跃的，应该重置定时器
			//不做任何处理，只是为了激活select，更新下面的定时器

		case <-time.After(time.Hour * 1):
			//已经超时
			//将当前user强制的关闭
			user.SendMsg("长时间不活跃，自动下线\n")
			//销毁使用的资源
			close(user.C)
			//关闭连接
			conn.Close()
			//退出当前的Handler
			return
		}
	}
}

func (server *Server) Start() {
	// Sprintf用于格式化字符串，但不打印到标准输出，而是返回一个格式化后的字符串
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen error: ", err)
		return
	}

	defer listener.Close()

	//启动监听Message的goroutine
	go server.ListenMessager()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listen accept error: ", err)
			continue
		}

		go server.Handler(conn)
	}
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}
