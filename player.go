package catchace

import (
	"encoding/json"
	"errors"
	"time"
)

type Player struct {
	username string      // 玩家名称
	out      chan []byte // 输出到客户端channel
	in       chan []byte // 从客户端读取channel
	cards    []string    // 已经抽到的卡
}

// 发送并等待回复
// msg 待发送消息
func (p *Player) Request(msg Msg) (resp Msg) {
	bins, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	// 发送
	p.out <- bins
	// 等待回复
	return p.WaitMsg()
}

// 发送消息，不需要回复
func (p *Player) Send(msg Msg) {
	go func() {
		bins, err := json.Marshal(msg)
		if err != nil {
			panic(err)
		}
		// 发送
		p.out <- bins
	}()
}

func (p *Player) SendMsg(action string, data interface{}) {
	p.Send(Msg{Username: p.username, Action: action, Data: data})
}

// 等待收到消息
func (p *Player) WaitMsg() (resp Msg) {
	reply := <-p.in
	err := json.Unmarshal(reply, &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

// 带有超时的请求
// 如果请求发送成功之后
func (p *Player) RequestTT(msg Msg, duration time.Duration) (resp Msg, err error) {
	bins, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	// 发送
	p.out <- bins

	timeout := make(chan bool)
	go func() {
		time.Sleep(duration)
		timeout <- true
	}()
	select {
	case reply := <-p.in:
		err := json.Unmarshal(reply, &resp)
		if err != nil {
			panic(err)
		}
		return resp, nil
	case <-timeout:
		// 超时
		return resp, errors.New("等待超时")
	}
}
