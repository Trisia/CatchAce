/*
     游戏状态：
     - Offline 没有连接服务器（用户可能没有创建）
     - Online 连接完成在大厅中
     - InRoom 在房间中
     - Gaming 游戏中
     - End - 游戏结束
     */
const GameStatus = {
    "Offline": "Offline",   //- Offline 没有连接服务器（用户可能没有创建）
    "Online": "Online",     //- Online 连接完成在大厅中
    "InRoom": "InRoom",     //- InRoom 在房间中
    "Gaming": "Gaming",     //- Gaming 游戏中
    "End": "End"            //- End - 游戏结束
};

/*
    游戏消息类型
 */
const Action = {
    ReqUseQ: "ReqUseQ",     // - 询问是否使用Q跳过抽卡
    RespUseQ: "RespUseQ",   // - 响应已经是否使用Q跳过抽卡
    UsedQ: "UsedQ",         // - 广播某玩家已经使用了Q跳过抽卡
    Drawing: "Drawing",     // - 玩家正在抽卡
    DoDraw: "DoDraw",       // - 玩家触发抽卡（客户端发起）
    Notice: "Notice",       // - 通知某玩家抽到卡
    RequestJ: "RequestJ",   // - 请求加酒
    AddSake: "AddSake",     // - 公告加酒数量
    Punish: "Punish",       // - 罚酒
    StartGame: "StartGame", // - 开始游戏
    EndGame: "EndGame",     // - 游戏结束
    OwnerExit: "OwnerExit", // - 房主退出了游戏
    GameInfo: "GameInfo",   // - 游戏状态广播
};

var app = new Vue({
    el: '#app',
    data: {
        username: "",
        conn: '',
        newRoomName: '',
        roomInterval: '',
        roomList: [],
        status: GameStatus.Offline,
        gameVal: {
            roomName: '',
            // 是否是房间的房主
            roomOwner: false,
            players: [], // 房间中玩家
            remain: 20, // 剩余卡牌数量
            sake: 1, // 加酒信息
            currentPlayerIndex: 0, // 但前正在操作玩家
            direction: 1,         // 轮转方向 1 表示顺序，-1 表示逆序
            hand: []          // 玩家当前手牌
        }
    },
    watch: {
        // 客户端段游戏状态监听
        status: function (newStatus, oldStatus) {
            console.log("Status:", oldStatus, "->", newStatus);
            if (oldStatus === newStatus) {
                return
            }
            // oldStatus: Gaming -> Online 启动拉取房间列表
            if (newStatus === GameStatus.Online) {
                this.pullRoomInterval()
            }
            // newStatus: any -> InRoom 取消房间列表拉取
            else if (newStatus === GameStatus.InRoom) {
                clearInterval(this.roomInterval)
            }
        }
    },
    methods: {
        // 开始/重开游戏
        restartGame: function () {
            axios.get('/CatchAce/start?roomName='
                + this.gameVal.roomName);
        },
        // 退出房间
        exit: function () {
            if (this.status === GameStatus.Offline
                || this.status === GameStatus.Online) {
                console.warn("未处于任何房间，无法退出");
                return;
            }
            // 从房间中移除
            axios.delete('/CatchAce/player?' +
                'roomName=' + this.gameVal.roomName +
                '&userName=' + this.username).finally(() => {
                // 回到大厅
                this.status = GameStatus.Online
            })
        },
        // 加入游戏
        join: function (room) {
            let data = new FormData();
            data.append('username', this.username);
            data.append('roomName', room.roomName);
            axios.post("/CatchAce/join", data).then(res => {
                // 加入房间
                this.status = GameStatus.InRoom;
            }).catch(error => {
                if (error.response) {
                    alert(error.response.data);
                }
            })
        },
        // 房间描述
        roomDesp: function (room) {
            let txt = "创建者: " + room.creator;
            txt += "\n当前玩家数量: " + room.playerNum;
            return txt
        },
        // 定期拉取房间列表
        pullRoomInterval: function () {
            // 调用时立刻拉取
            this.pullRoom();
            // 如果已经存在一个定时器，那么删除
            if (this.roomInterval) {
                clearInterval(this.roomInterval)
            }
            // 定期拉取游戏房间列表
            this.roomInterval = setInterval(() => {
                this.pullRoom();
            }, 5000)
        },
        // 拉取房间列表
        pullRoom: function () {
            axios.get('/CatchAce')
                .then((response) => {
                    this.roomList = (!response.data) ? [] : response.data;
                })
                .catch((error) => {
                    console.log(error.response);
                });
        },
        // 创建房间
        createRoom: function () {
            if (!this.newRoomName) {
                console.log("房间名不能为空");
                return;
            }
            let data = new FormData();
            data.append('username', this.username);
            data.append('roomName', this.newRoomName);
            axios.post("/CatchAce/create", data).then(res => {
                this.status = GameStatus.InRoom;
                this.gameVal.roomName = this.newRoomName;
                this.gameVal.roomOwner = true;
            }).catch(error => {
                if (error.response) {
                    alert(error.response.data);
                }
            })
        },
        // 检查并连接服务器
        checkAndConnect: function () {
            if (!this.username) {
                alert("必须填写用户名");
                return
            }
            console.log(">> Set username:", this.username);
            localStorage["username"] = this.username;
            axios.get('/userExist?username=' + this.username).then(res => {
                if (res.data) {
                    alert("用户名: " + this.username + " 已存在，换一个试试吧");
                    return
                }
                this.connect()
            });
        },
        connect: function () {
            if (!window["WebSocket"]) {
                alert("您的浏览器不支持！");
                return
            }
            this.conn = new WebSocket("ws://" + document.location.host + "/cnn");
            this.conn.onclose = this.onClose;
            this.conn.onmessage = this.onMsg;
            this.conn.onopen = this.onOpen;
            this.status = GameStatus.Online;
            // 开始定期拉取
            this.pullRoomInterval();
        },
        // 发送消息
        send: function(msg){
            if (this.conn) {
                this.conn.send(JSON.stringify(msg));
            }
        },
        onClose: function (evt) {
            this.status = GameStatus.Offline;
            console.warn("<< Connection close:", evt);
        },
        onMsg: function (evt) {
            // 序列化消息
            let msg = JSON.parse(evt.data);
            this.messageProcess(msg)
        },
        // 消息处理器
        messageProcess: function (msg) {
            switch (msg.Action) {
                // 同步游戏状态
                case Action.GameInfo:
                    this.renewGameInfo(msg.Data);
                    break;
                // 开始游戏
                case Action.StartGame:
                    this.status = GameStatus.Gaming;
                    break;
                // 抽卡指令
                case Action.DoDraw:
                    this.draw(msg.Username);
                    break;
                default:
                    console.log(">>", msg);
            }
        },
        // 抽卡或显示玩家正在抽卡
        draw: function (username) {
            if (username === this.username) {
                // 自己抽卡
                alert("点击确定抽卡");
                this.send({
                    Action: Action.DoDraw
                })
            } else {
                // TODO 显示某玩家正在抽卡
                console.log(">> 玩家:", this.username, "正在抽卡...")
            }
        },

        // 更新游戏信息
        renewGameInfo: function (data) {
            this.gameVal.sake = data.Sake;
            this.gameVal.currentPlayerIndex = data.CurrentPlayerIndex;
            this.gameVal.direction = data.Direction;
            this.gameVal.remain = data.RemainCard.length;
            this.$set(this.gameVal, 'players', data.Players);
            // 设置玩家自己的手牌
            for (let p in data.Players) {
                if (p.Username === this.username) {
                    this.$set(this.gameVal, "hand", data.Cards);
                    break;
                }
            }
        },
        onOpen: function (evt) {
            this.conn.send(this.username);
            console.log(">>", this.username, "Connected")
        },
    },
    mounted: function () {
        this.username = localStorage["username"];
    },
    destroyed: function () {
        if (this.conn) {
            this.conn.close()
        }
    }
});
