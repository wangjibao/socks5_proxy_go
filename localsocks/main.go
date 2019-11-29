package main

import (
	"fmt"
	"log"
	"mysocks"
	"net"
)

func main() {
	log.SetFlags(log.Lshortfile)

	config := &mysocks.Config{}
	config.ReadConfigFromFile() // 会覆盖掉上面的 LocalListenAddr 的赋值
	// config.WriteConfigToFile()
	// 启动 local 端，并监听
	lsLocal, err := mysocks.NewListenLocal(config.Password, config.LocalListenAddr, config.RemoteServerAddr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Fatalln(lsLocal.Listen(func(listenAddr net.Addr) {
		log.Println(fmt.Sprintf("mysocks-local 启动成功，配置如下：\n本地监听地址：%s\t 远程服务地址：%s\t 密码：%s", listenAddr, config.RemoteServerAddr, config.Password))
	}))
}
