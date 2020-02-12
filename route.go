package catchace

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func RouteConfig(r *gin.Engine) {
	r.StaticFS("/game", http.Dir("static"))
	// 配置Websocket 连接，与客户端建立连接
	r.GET("/cnn", func(c *gin.Context) {
		useConnect(c.Writer, c.Request)
	})
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/game/index.html")
	})
	// 摸A游戏路由
	//
	r.POST("/CatchAce/create", createCatchAce)
	// 获取房间列表
	r.GET("/CatchAce", listCatchAce)
	// 加入房间
	r.POST("/CatchAce/join", joinCatchAce)
	// 退出游戏
	r.DELETE("/CatchAce/player", catchAceExit)
	// 开始/重启 游戏
	r.GET("/CatchAce/start", catchAceStart)
	// 检查用户是否已经存在
	r.GET("/userExist", func(c *gin.Context) {
		username := c.Query("Username")
		use := OnlineUses[username]
		if use == nil {
			c.String(http.StatusOK, "false")
			return
		}
		c.String(http.StatusOK, "true")
	})
}
