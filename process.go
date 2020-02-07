package main

import (
	"strings"
	"time"
)

// processCard 处理抽到的卡
// return true 表示游戏结束， false表示游戏继续
func (r *CatchAce) processCard(card string, p *Player) bool {
	// 取出抽到卡的数组：10，J，Q，K，A
	key := strings.Split(card, ",")[0]
	r.counter[key] += 1
	switch key {
	case "10":
		// 增长计数取负值
		r.incBase = -r.incBase
	case "J":
		// 加酒
		reqAddAke(r, p)
	case "Q":
		return false
	case "K":
		selfDrink(r, p)
	case "A":
		return ace(r, p)
	}
	return false
}

// 抽到A判断是否结束游戏
func ace(r *CatchAce, p *Player) bool {
	cnt := r.counter["A"]
	if cnt == 4 {
		r.Broadcast(Msg{
			Username: p.username,
			Action:   "Punish",
			Data:     r.counter["K"],
		})
		return true
	}
	return false
}

// 自罚
func selfDrink(r *CatchAce, p *Player) {
	// 广播喝酒信息
	r.Broadcast(Msg{
		Username: p.username,
		Action:   "Punish",
		Data:     r.counter["K"],
	})
}

// 请求加酒
func reqAddAke(r *CatchAce, p *Player) {
	// 向指定的玩家发送加酒请求
	// 等待玩家响应
	resp, err := p.RequestTT(Msg{
		Username: p.username,
		Action:   "RequestJ",
	}, 5*time.Second)
	added := 1
	if err == nil {
		// 如果玩家没有处理没有超时，那么取得 请求加酒值。
		num, ok := resp.Data.(int)
		if ok {
			added = num
		}
	}
	r.sake += added
	// 广播加酒信息
	r.Broadcast(Msg{
		Username: p.username,
		Action:   "AddSake",
		Data:     added,
	})
}
