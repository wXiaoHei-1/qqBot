package main

import (
	"context"
	"log"
	"os"
	"qqbot/common/clients"
	"qqbot/common/service"
	"qqbot/common/types"
	"qqbot/server"
	"qqbot/utils"
	"strings"
	"time"
)

var (
	ctx          context.Context
	httpClient   *service.HttpClient
	ws           *types.WebsocketAP
	replyMessage string
	finishOrNot  bool
	err          error
	timer        *time.Timer
)

func init() {
	// 读取配置信息
	utils.NewConfig()
	// 初始化数据库链接
	clients.NewDBClient(utils.ConfigInfo)
	// 初始化成语库
	server.NewIdiomMap()
	// 初始化http连接
	httpClient = service.NewClient(utils.ConfigInfo.AppID, utils.ConfigInfo.Token, 3*time.Second)
}

func main() {
	// 获取context
	ctx = context.Background()
	// 通过http获取webSocket连接地址
	ws, err = httpClient.GetWSS(ctx)
	if err != nil {
		log.Fatalln("websocket err:", err)
		os.Exit(1)
	}
	log.Printf("%+v, err:%v", ws, err)

	// 注册@消息的回调函数
	var atMessage service.ATMessageEventHandler = AtMessageEventHandler
	intent := service.RegisterHandlers(atMessage)
	err = service.NewSessionManager().Start(ws,
		&service.Token{
			AppID:       utils.ConfigInfo.AppID,
			AccessToken: utils.ConfigInfo.Token,
			Type:        "Bot",
		}, intent)
	if err != nil {
		log.Printf("Failed to start session manager for appID: %d with error: %v", utils.ConfigInfo.AppID, err)
	}
}

// AtMessageEventHandler 处理 @机器人消息的回调函数
func AtMessageEventHandler(event *types.WSPayload, data *types.Message) error {
	messageContent := data.Content[strings.Index(data.Content, ">")+2:]
	// FinishOrNot游戏是否还在进行标记位
	if finishOrNot {
		// 重置定时器
		resetTimer(data)
		replyMessage = GameInProgress(messageContent, data)
	} else {
		replyMessage = InitialOperation(messageContent, data)
	}
	_, err = httpClient.PostMessage(ctx, data.ChannelID, &types.MessageToCreate{MsgID: data.ID, Content: replyMessage})
	if err != nil {
		log.Println("Failed to post message to channel:", data.ChannelID, "with message:", replyMessage, "and error:", err)
	}
	return nil
}

// GameInProgress 游戏还在进行中
func GameInProgress(messageContent string, data *types.Message) string {
	// 游戏还在进行中，输入/成语接龙则认为用户希望重新开始游戏
	if strings.EqualFold(messageContent, "/成语接龙") {
		server.ResetCurrentIdiom()
		resetTimer(data)
		return "好的游戏重新开始，请说出一个四字成语。"
	}
	// 游戏还在进行中，输入/quit则退出游戏
	if strings.EqualFold(messageContent, "/quit") {
		finishOrNot = false
		stopTimer()
		server.ResetCurrentIdiom()
		return "好的,游戏结束"
	}
	// flag表示是否需要结束游戏，词库没有与用户输入匹配的词语则结束游戏
	interlocking, flag := server.ChengYvInterlocking(messageContent)
	if flag {
		finishOrNot = false
		stopTimer()
		return interlocking
	}
	return interlocking
}

// InitialOperation 初始状态下的操作
func InitialOperation(messageContent string, data *types.Message) string {
	// 输入指令/成语接龙开始游戏，并将游戏记号位标为正在进行true
	if strings.EqualFold(messageContent, "/成语接龙") {
		finishOrNot = true
		resetTimer(data)
		return "欢迎来到成语接龙游戏！请说出第一个四字成语"
	}
	// 因为当前没有任何进度，需要提醒用户当前并没有进行游戏
	if strings.EqualFold(messageContent, "/quit") {
		return "当前没有进行游戏"
	}
	// 指令之外的消息，认为是与用户之间的对话
	reply := server.SendMessage(messageContent, utils.ConfigInfo.DashScopeAPIKey)
	return reply
}

// resetTimer 函数用于重置游戏计时器,并在计时器超时时执行相应的结束游戏操作
func resetTimer(data *types.Message) {
	if timer != nil {
		timer.Stop()
	}
	timer = time.NewTimer(60 * time.Second)
	go func() {
		<-timer.C
		// 60秒内没有回答,结束游戏
		finishOrNot = false
		server.ResetCurrentIdiom()
		_, _ = httpClient.PostMessage(ctx, data.ChannelID, &types.MessageToCreate{Content: "60秒内没有回答,游戏结束。"})
	}()
}

// stopTimer 函数用于停止当前正在运行的游戏计时器
func stopTimer() {
	if timer != nil {
		timer.Stop()
	}
}
