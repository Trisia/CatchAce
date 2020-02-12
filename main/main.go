package main

import (
	"catchace"
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

var port = flag.String("port", "80", "服务监听端口")

func main() {
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	catchace.RouteConfig(r)
	log.Printf(">> 服务端口:%s", *port)
	r.Run(":" + *port)
}

func report() {
	go func() {
		ticker := time.NewTicker(10 * time.Second).C
		for {
			<-ticker
			log.Printf("当前在线人数: %d", len(catchace.OnlineUses))
		}
	}()
}
