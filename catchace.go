package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

var QPattern = regexp.MustCompile(`Q,.`)

type CatchAce struct {
	name               string
	cards              []string
	manager            *Player
	players            []*Player
	incBase            int            // 增加基数
	currentPlayerIndex int            // 但前玩家索引
	counter            map[string]int // 扑克计数器
	sake               int            // 加酒的数量 (日语　さけ)
	status             string         // 游戏状态
}

type Msg struct {
	Username string
	Action   string // 动作名称
	Data     string // 动作数据
}

// 创建一个新的游戏房间
func NewCatchAce(name string, manager *Player) *CatchAce {
	room := CatchAce{
		name:    name,
		manager: manager,
	}
	room.players = append(room.players, manager)
	room.Init()
	return &room
}

// Join 增加玩家数量
func (r *CatchAce) Join(p *Player) {
	if p == nil {
		return
	}
	r.players = append(r.players, p)
}

// Exit 玩家退出游戏
func (r *CatchAce) Exit(username string) {
	for i, player := range r.players {
		if player.username == username {
			r.players = append(r.players[0:i], r.players[i+1:]...)
			break
		}
	}
}

// Init  游戏初始化
func (r *CatchAce) Init() {
	// 洗牌
	r.shuffleCard()
	// 增长计数变回1
	r.incBase = 1
	r.currentPlayerIndex = -1
	r.counter = make(map[string]int)
	r.sake = 0
	r.status = "wait"
}

// 重新洗牌
func (r *CatchAce) shuffleCard() {
	colors := []string{
		"S", // 黑桃:S-Spade
		"H", // 红桃:H-Heart
		"C", // 梅花:C-Club
		"D", // 方块:D-Diamond
	}
	cardNumbers := []string{"10", "J", "Q", "K", "A"}
	// 清空已有卡
	r.cards = []string{}
	for _, num := range cardNumbers {
		for _, color := range colors {
			r.cards = append(r.cards, fmt.Sprintf("%s,%s", num, color))
		}
	}
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano() - rand.Int63n(100))
	rand.Shuffle(len(r.cards), func(i, j int) {
		r.cards[i], r.cards[j] = r.cards[j], r.cards[i]
	})
}

// 启动游戏
func (r *CatchAce) Run() {
	r.status = "gaming"
	for {
		// 1. 选择出待抽卡的玩家
		nextPlayer := r.choosePlayer()
		// 2. 玩家是否跳过抽卡阶段
		if r.skipDrawCard(nextPlayer) {
			// 下一个玩家抽卡
			continue
		}
		// 3. 通知当前玩家正在抽卡
		r.Broadcast(Msg{
			Username: nextPlayer.username,
			Action:   "Drawing",
		})
		// 4. 等待玩家确定抽卡
		nextPlayer.WaitMsg()
		// 5. 抽卡
		card := r.drawCard(nextPlayer)
		// 6. 广播抽到的卡
		r.Broadcast(Msg{
			Username: nextPlayer.username,
			Action:   "Notice",
			Data:     card,
		})
		// 7. 处理抽到的卡。
		endOfGame := r.processCard(card, nextPlayer)
		if endOfGame {
			if r.sake == 0 {
				r.sake = 1
			}

			r.Broadcast(Msg{
				Username: nextPlayer.username,
				Action:   "EndOfGame",
				Data:     strconv.Itoa(r.sake),
			})
			r.status = "End"
			break
		}
	}
}

// 玩家抽卡
// 返回抽到的卡
func (r *CatchAce) drawCard(p *Player) string {
	// 取出牌面的第一个元素
	newCard := r.cards[0]
	r.cards = r.cards[1:]
	p.cards = append(p.cards, newCard)
	return newCard
}

// 选择下一位抽卡玩家
func (r *CatchAce) choosePlayer() *Player {
	r.currentPlayerIndex += r.incBase
	if r.currentPlayerIndex < 0 {
		r.currentPlayerIndex = len(r.players) - 1
	} else {
		r.currentPlayerIndex %= len(r.players)
	}
	return r.players[r.currentPlayerIndex]
}

// 询问是否使用跳过卡
// return true 表示使用 false - 没有或不使用
func (r *CatchAce) skipDrawCard(p *Player) bool {
	useCard := ""
	if p.cards != nil {
		for i, card := range p.cards {
			// 从玩家已经拥有的卡组中寻找未使用的的Q
			if QPattern.MatchString(card) && len(card) < 4 {
				useCard = p.cards[i]
				// 标记已经被使用
				p.cards[i] = p.cards[i] + ",used"
				break
			}
		}
	}
	if useCard == "" {
		// 不存在Q
		return false
	}
	// 询问是否使用Q
	resp := p.Request(Msg{
		Username: p.username,
		Action:   "ReqUseQ",
	})
	parseBool, err := strconv.ParseBool(resp.Data)
	if err != nil {
		panic(err)
	}
	if parseBool {
		// 广播某人使用Q
		r.Broadcast(Msg{
			Username: p.username,
			Data:     "UsedQ",
		})
	}
	return parseBool
}

// 广播消息
func (r *CatchAce) Broadcast(msg Msg) {
	for _, p := range r.players {
		p.Send(msg)
	}
}
