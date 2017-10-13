package main

import (
	"fmt"
	"log"
	"net"

	"flag"

	"../../cmd"
	"../../core"
	"../../local"
)

const (
	DefaultListenAddr = ":7448"
)

var version = "master"

func main() {
	log.SetFlags(log.Lshortfile)

	passwd := flag.String("password", "", "")
	rmt := flag.String("remote", "", "")
	listen := flag.String("listen", DefaultListenAddr, "")
	flag.Parse()

	if *rmt == "" || *passwd == "" {
		config := &cmd.Config{
			ListenAddr: DefaultListenAddr,
		}
		config.ReadConfig()
		config.SaveConfig()

		*rmt = config.RemoteAddr
		*passwd = config.Password
		*listen = config.ListenAddr
	}

	password, err := core.ParsePassword(*passwd)
	if err != nil {
		log.Fatalln(err)
	}
	listenAddr, err := net.ResolveTCPAddr("tcp", *listen)
	if err != nil {
		log.Fatalln(err)
	}
	remoteAddr, err := net.ResolveTCPAddr("tcp", *rmt)
	if err != nil {
		log.Fatalln(err)
	}

	// 启动 local 端并监听
	lsLocal := local.New(password, listenAddr, remoteAddr)
	lsLocal.AfterListen = func(listenAddr net.Addr) {
		log.Printf("lightsocks-local:%s 启动成功 监听在 %s\n", version, listenAddr.String())
		log.Println("使用配置：", fmt.Sprintf(`
本地监听地址 listen：
%s
远程服务地址 remote：
%s
密码 password：
%s
	`, listenAddr, remoteAddr, password))
	}
	log.Fatalln(lsLocal.Listen())
}
