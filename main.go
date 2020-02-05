package main

import (
	"github.com/gin-gonic/gin"
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
