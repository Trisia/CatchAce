package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)
// 创建摸A房间
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

// 列出所有摸A的房间
func listCatchAceRoom() func(c *gin.Context) {
	return func(c *gin.Context) {
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
}
