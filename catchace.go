package catchace

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
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
	status             string         // 游戏状态: wait - 等待、gaming - 游戏中、end - 结束、close - 关闭
}

type Msg struct {
	Username string
	Action   string      // 动作名称
	Data     interface{} // 动作数据
}

// 玩家临时数据
type playerDto struct {
	Username string
	Cards    []string
}

// 当前的游戏信息
type gameInfoDto struct {
	Players            []playerDto // 游戏中玩家信息
	RemainCard         []string    // 剩余卡牌
	Sake               int         // 已经增加的酒信息
	CurrentPlayerIndex int         // 但前正在操作玩家
	Direction          int         // 轮转方向 1 表示顺序，-1 表示逆序
}

// 创建一个新的游戏房间
func NewCatchAce(name string, manager *Player) *CatchAce {
	room := CatchAce{
		name:    name,
		manager: manager,
	}
	room.players = append(room.players, manager)
	room.Init()
	// 广播游戏信息
	room.GameInfoBroadcast()
	return &room
}

// getGameInfo 获取但前游戏状态
func (r *CatchAce) getGameInfo() gameInfoDto {
	var players []playerDto
	for _, p := range r.players {
		players = append(players, playerDto{
			Username: p.username,
			Cards:    p.cards,
		})
	}
	return gameInfoDto{
		Players:            players,
		RemainCard:         r.cards,
		Sake:               r.sake,
		CurrentPlayerIndex: r.currentPlayerIndex,
		Direction:          r.incBase,
	}
}

// 广播游戏信息
func (r *CatchAce) GameInfoBroadcast() {
	tick := time.Tick(500 * time.Millisecond)
	go func() {
		for r.status != "close" {
			<-tick
			r.Broadcast(Msg{
				Action: "GameInfo",
				Data:   r.getGameInfo(),
			})
		}
	}()
}

func (r *CatchAce) Close() {
	r.status = "close"
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
	r.sake = 1
	r.status = "wait"
	for _, p := range r.players {
		// 清空玩家所有手牌
		p.cards = []string{}
	}
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
			r.Broadcast(Msg{
				Username: nextPlayer.username,
				Action:   "EndOfGame",
				Data:     r.sake,
			})
			r.status = "wait"
			break
		}
		time.Sleep(5 * time.Second)
	}
}

// 玩家抽卡
// 返回抽到的卡
func (r *CatchAce) drawCard(p *Player) string {
	// 从剩余扑克中随机取出一个
	index := rand.Intn(len(r.cards))
	newCard := r.cards[index]
	r.cards = append(r.cards[0:index], r.cards[index+1:]...)
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
	index := -1
	for i, card := range p.cards {
		// 从玩家已经拥有的卡组中寻找未使用的的Q
		if QPattern.MatchString(card) && len(card) < 4 {
			index = i
			break
		}
	}
	if index == -1 {
		// 不存在Q
		return false
	}
	// 询问是否使用Q
	resp, err := p.RequestTT(Msg{
		Username: p.username,
		Action:   "ReqUseQ",
	}, 10*time.Second)
	if err != nil {
		log.Printf("玩家:%s 是否使用Q超时...", p.username)
		// 等待超时默认不使用
		return false
	}
	use, ok := resp.Data.(bool)
	if ok && use {
		// 表示卡已经使用
		p.cards[index] = p.cards[index] + ",used"
		// 广播某人使用Q
		r.Broadcast(Msg{
			Username: p.username,
			Action:   "UsedQ",
			Data:     p.cards[index],
		})
	}
	return use
}

// 广播消息
func (r *CatchAce) Broadcast(msg Msg) {
	for _, p := range r.players {
		p.Send(msg)
	}
}
