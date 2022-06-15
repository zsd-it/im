package main

import (
	"fmt"
	"net"
)

type ImServer struct {
	Ip               string
	Port             int
	OnlineMap        map[string]*UserServer
	BroadcastChannel chan string
}

//开启服务
func (im ImServer) Start() {
	//创建一个监听
	var ipStr = fmt.Sprintf("%s:%d", im.Ip, im.Port)
	listener, err := net.Listen("tcp", ipStr)

	if err != nil {
		fmt.Printf("监听失败：%s, err: %v", ipStr, err)
		return
	}

	//关闭连接
	defer listener.Close()

	//开启全局频道
	go im.broadcast()

	fmt.Println("成功启动Im...")

	for {
		//等待连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("监听失败：%s, err: %v", ipStr, err)
			continue
		}

		//处理连接
		go im.Handle(conn)
	}
}

func (im *ImServer) Handle(conn net.Conn) {
	//用户处理
	userServer := newUserServer(conn, im)
	userServer.Handle()
}

//广播通道
func (im ImServer) broadcast() {
	//循环获取消息, 然后发给所有人
	for msg := range im.BroadcastChannel {
		//不考虑顺序问题
		go func() {
			for _, user := range im.OnlineMap {
				user.UChannel <- msg

				fmt.Println(user.Name)
			}
		}()
	}
}

func (im ImServer) sendBroadcast(str string) {
	im.BroadcastChannel <- str
}

func NewImServer() *ImServer {
	return &ImServer{
		Ip:               "127.0.0.1",
		Port:             8888,
		BroadcastChannel: make(chan string),
		OnlineMap:        make(map[string]*UserServer),
	}
}
