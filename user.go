package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

func (user *User) findUserOnline() {

}

// Online 用户的上线业务
func (user *User) Online() {
	//用户上线，将用户加入到OnlineMap中
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	// 广播当前用户上线的消息
	user.server.Broadcast(user, "已上线")
}

// Offline 用户的下线业务
func (user *User) Offline() {
	//用户上线，将用户加入到OnlineMap中
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	// 广播当前用户下线的消息
	user.server.Broadcast(user, "下线")

}

// 给当前user对应的客户端发消息
func (user User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户有哪些
		user.server.mapLock.Lock()
		for _, u := range user.server.OnlineMap {
			onlineMsg := "[" + u.Addr + "]" + u.Name + "在线...\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式：rename|张三
		newName := strings.Split(msg, "|")[1]

		//判断name是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("用户名已存在\n")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()
			user.Name = newName
			user.SendMsg("您已更新用户名:" + user.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式：to|张三|消息内容

		//1.获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("消息格式不正确，请使用\"to|张三|消息内容\"格式。\n")
			return
		}

		//2.根据用户名得到对方的user对象
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("该用户名不存在\n")
			return
		}
		//获取消息内容，通过对方的user对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMsg(user.Name + "对您说:" + content + "\n")
	} else {
		user.server.Broadcast(user, msg)
	}
}

// ListenMessage 监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}
