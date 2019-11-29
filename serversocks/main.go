package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"mysocks"
)

func main() {
	log.SetFlags(log.Lshortfile)
	// 获取服务端监听端口
	// port, err := strconv.Atoi(os.Getenv("MYSOCKS_SERVER_PORT"))
	// if err != nil {
	// 	port = 31415
	// } 
	config := &mysocks.Config{}
	config.ReadConfigFromFile()
	// 启动 server 端并监听
	lsServer, err := mysocks.NewListenServer(config.Password, config.LocalListenAddr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Fatalln(lsServer.Listen(func(listenAddr net.Addr) {
		log.Println(fmt.Sprintf("mysocks-server 启动成功，配置如下：\n 服务端监听地址：%s\t 密码：%s\n", listenAddr, config.Password))
	}))
}
