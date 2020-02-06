package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

// 在线用户
var OnlineUses = make(map[string]*OnlineUse)

// 当前的游戏
var Games = make(map[string]*CatchAce)

func main() {
	r := gin.Default()
	routeConfig(r)
	r.Run(":80")
}

func report() {
	go func() {
		ticker := time.NewTicker(10 * time.Second).C
		for {
			<-ticker
			log.Printf("当前在线人数: %d", len(OnlineUses))
		}
	}()
}
