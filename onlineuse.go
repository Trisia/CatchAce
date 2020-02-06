package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type OnlineUse struct {
	// websocket connection.
	conn *websocket.Conn
	// 用户名，由连接建立后第一条消息决定
	username string
	// 输出到客户端channel
	out chan []byte
	// 从客户端读取channel
	in chan []byte
}

func (o *OnlineUse) Player() *Player {
	return &Player{
		username: o.username,
		in:       o.in,
		out:      o.out,
	}
}

// readPump 不断监听从客户端收到的消息
// 并且传递给对应的channel
func (c *OnlineUse) readPump() {
	defer func() {
		c.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Printf(">> [%s]: %s", c.username, message)
		go func() { c.in <- message }()
	}
}

// writePump 持续监听消息
func (c *OnlineUse) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()
	for {
		select {
		case message, ok := <-c.out:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("write:", err)
				return
			}
			log.Printf("<< [%s]: %s", c.username, message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 删除玩家所拥有的房间
func (c *OnlineUse) DeleteRoom() {
	for roomName, room := range Games {
		// 找到该用户创建的所有房间
		if room.manager.username == c.username {
			log.Printf(">> 因为房主: [%s]退出，删除房间: [%s]", c.username, roomName)
			delete(Games, roomName)
			// 通知其他玩家房主下线，房间解散
			for _, player := range room.players {
				if player.username == c.username {
					// 跳过房主
					continue
				}
				player.Send(Msg{Action: "OwnerExit"})
			}
		}
	}
}

// Close 客户端关闭
func (c *OnlineUse) Close() {
	log.Printf(">> Use [%s] Disconnect.", c.username)
	// 从在线用户列表中注销
	delete(OnlineUses, c.username)
	// 删除用户创建的房间
	c.DeleteRoom()
	c.conn.Close()
}

// 用户连接
func useConnect(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// 打开连接的同时等待第一条消息
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}
	// 第一条消息携带上用户名
	username := string(msg)
	log.Printf("<< Use [%s] connected.", username)

	client := OnlineUse{
		conn:     conn,
		username: username,
		out:      make(chan []byte),
		in:       make(chan []byte),
	}
	// 加入在线用户列表
	OnlineUses[username] = &client

	go client.readPump()
	go client.writePump()
}
