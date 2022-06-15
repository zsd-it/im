package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type UserServer struct {
	Name     string
	Ip       string
	UChannel chan string
	Conn     net.Conn
	ImServer *ImServer
}

//接受消息
func (userServer *UserServer) getMsg() {
	for {
		msg := <-userServer.UChannel

		userServer.receiveMsg(msg)
	}
}

//默认执行
func (userServer *UserServer) Handle() {
	//消息监听
	go userServer.getMsg()

	//上线提醒
	userServer.onLine()

	//用户发消息处理
	userServer.sendMsg()
}

//上线
func (userServer *UserServer) onLine() {
	//加入在线列表
	userServer.ImServer.OnlineMap[userServer.Name] = userServer
	userServer.ImServer.sendBroadcast(fmt.Sprintf("欢迎用户%s进入直播间\n", userServer.Name))
}

//下线
func (userServer *UserServer) offLine() {
	delete(userServer.ImServer.OnlineMap, userServer.Name)
	userServer.ImServer.sendBroadcast(fmt.Sprintf("%s 下线\n", userServer.Name))
}

//发送消息
func (userServer *UserServer) sendMsg() {
	var isLive = make(chan int)
	//监控用户发送的消息
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := userServer.Conn.Read(buf)

			if n == 0 {
				userServer.offLine()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn read error :", err)
			}

			//去除\n
			msg := string(buf[:n-1])
			userServer.DoMsg(msg)

			isLive <- 0
		}
	}()

	//10s 自动下线
	for {
		select {
		case <-isLive:
		//用户10s 内有操作
		case <-time.After(time.Second * 10):
			userServer.receiveMsg("自动下线！")
			//10s超时
			_ = userServer.Conn.Close() //关闭连接

			close(userServer.UChannel) //关闭通道

			return
		}
	}
}

//处理接受到的消息
func (userServer *UserServer) DoMsg(msg string) {
	if msg == "who" {
		userServer.OrderWho()
	} else {
		msg = fmt.Sprintf("%s:%s\n", userServer.Name, msg)
		userServer.ImServer.sendBroadcast(msg)
	}
}

//接收消息
func (userServer *UserServer) receiveMsg(msg string) {
	_, _ = userServer.Conn.Write([]byte(msg))
}

//指令 who
func (userServer *UserServer) OrderWho() {
	var whoMsg string
	for _, user := range userServer.ImServer.OnlineMap {
		whoMsg += fmt.Sprintf("%s 在线。。。\n", user.Name)
	}

	userServer.receiveMsg(whoMsg)
}

func newUserServer(conn net.Conn, im *ImServer) *UserServer {
	userAddress := conn.RemoteAddr().String()
	return &UserServer{
		Name:     userAddress,
		Ip:       userAddress,
		UChannel: make(chan string),
		Conn:     conn,
		ImServer: im,
	}
}
