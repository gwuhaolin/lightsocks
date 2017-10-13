package main

import (
	"log"
	"net"
	"fmt"
	"../../server"
	"../../cmd"
	"../../core"
	"github.com/phayes/freeport"
)

var version = "master"

func main() {
	log.SetFlags(log.Lshortfile)

	var err error

	// 服务端监听端口随机生成
	port, err := freeport.GetFreePort()
	if err != nil {
		// 随机端口失败就采用 7448
		port = 7448
	}
	config := &cmd.Config{
		ListenAddr: fmt.Sprintf(":%d", port),
		// 密码随机生成
		Password: core.RandPassword().String(),
	}
	config.ReadConfig()
	config.SaveConfig()

	password, err := core.ParsePassword(config.Password)
	if err != nil {
		log.Fatalln(err)
	}
	listenAddr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
	if err != nil {
		log.Fatalln(err)
	}

	// 启动 server 端并监听
	lsServer := server.New(password, listenAddr)
	lsServer.AfterListen = func(listenAddr net.Addr) {
		log.Printf("lightsocks-server:%s 启动成功 监听在 %s\n", version, listenAddr.String())
		log.Println("使用配置：", fmt.Sprintf(`
本地监听地址 listen：
%s
密码 password：
%s
	`, listenAddr, password))
	}
	log.Fatalln(lsServer.Listen())
}
