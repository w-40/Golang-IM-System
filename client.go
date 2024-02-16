package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //当前client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	// 链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	// 返回对象
	return client
}

// DealResponse 处理server回应的消息，直接显示到标准输出即可
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听即可
	/** 这句话等价于
	for {
		buf := make()
		client.conn.Read(buf)
		fmt.Println(buf)
	}
	*/
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>请输入合法的数字<<<<<")
		return false
	}
}

// SelectUsers 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error: ", err)
	}
}

// PrivateChat 私聊模式
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println("请输入聊天对象的[用户名],exit退出")
	fmt.Scanln(&remoteName)
	for remoteName != "exit" {
		fmt.Println(">>>>>请输入消息内容,exit退出")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			//消息不为空，发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write error:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>>请输入聊天内容,输入exit退出")
			fmt.Scanln(&chatMsg)
		}
	}
}

// PublishChat 公聊模式
func (client *Client) PublishChat() {
	var chatMsg string
	fmt.Println(">>>>>请输入聊天内容,输入exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发给服务器
		//消息不为空，发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write error:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>>请输入聊天内容,输入exit退出")
		fmt.Scanln(&chatMsg)
	}

}

func (client *Client) updateName() bool {
	fmt.Println(">>>>>请输入用户名：")
	fmt.Scanln(&client.Name)
	senMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(senMsg))
	if err != nil {
		fmt.Println("conn Write error: ", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		// 根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			//公聊模式
			client.PublishChat()
			break
		case 2:
			//私聊模式
			client.PrivateChat()
			break
		case 3:
			//更新用户名
			client.updateName()
			break

		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "192.168.3.104", "设置服务器IP地址（默认是192.168.3.104）")
	flag.IntVar(&serverPort, "port", 8887, "设置服务器端口（默认是8888）")
}

func main() {
	//命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>链接服务器失败...")
		return
	}
	//单独开启一个goroutine去处理server的回执消息
	go client.DealResponse()
	fmt.Println(">>>>>>>链接服务器成功...")

	//启动客户端的业务
	client.Run()
}
