package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// createCatchAce 创建摸A房间
func createCatchAce(c *gin.Context) {
	username := c.PostForm("username")
	roomName := c.PostForm("roomName")

	user := OnlineUses[username]
	if user == nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("用户: [%s] 不存在"))
		return
	}
	room := Games[roomName]
	if room != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("房间: [%s] 已经存在，名称重复"))
		return
	}
	// 创建新的房间
	room = NewCatchAce(roomName, user.Player())
	log.Printf("用户: [%s] 创建了房间 [%s]", username, roomName)
	Games[roomName] = room
	c.String(http.StatusOK, "")
}

// listCatchAce 列出所有摸A的房间
func listCatchAce(c *gin.Context) {
	res := make(map[string]map[string]interface{})
	for roomName, game := range Games {
		res[roomName] = map[string]interface{}{
			"creator":   game.manager.username,
			"status":    game.status,
			"playerNum": len(game.players),
		}
	}
	c.JSON(http.StatusOK, res)
}

// joinCatchAce 加入房间
func joinCatchAce(c *gin.Context) {
	roomName := c.PostForm("roomName")
	userName := c.PostForm("userName")
	room := Games[roomName]
	if room == nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("房间 [%s] 不存在，无法加入房间", roomName))
		return
	}
	user := OnlineUses[userName]
	if user == nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("玩家 [%s] 不存在，无法加入房间", userName))
	}
	// 加入房间
	room.Join(user.Player())
	log.Printf("玩家: [%s] 加入房间 [%s]", userName, roomName)
}

// catchAceExit 玩家退出房间
func catchAceExit(c *gin.Context) {
	roomName := c.Query("roomName")
	userName := c.Query("userName")

	room := Games[roomName]
	if room.manager.username == userName {
		delete(Games, roomName)
	} else {
		room.Exit(userName)
	}
	c.String(http.StatusOK, "")

}

// catchAceStart 开始/重启 游戏
func catchAceStart(c *gin.Context) {
	roomName := c.Query("roomName")
	room := Games[roomName]
	log.Printf("房间: [%s] CatchA 开始游戏...", roomName)
	room.Init()
	go room.Run()
}
