package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func routeConfig(r *gin.Engine) {
	r.StaticFS("/game", http.Dir("static"))
	// 配置Websocket 连接，与客户端建立连接
	r.GET("/cnn", func(c *gin.Context) {
		useConnect(c.Writer, c.Request)
	})
	// 摸A游戏的
	//
	r.POST("/CatchAce/create", createCatchAce)
	// 获取房间列表
	r.GET("/CatchAce", listCatchAceRoom())
}
