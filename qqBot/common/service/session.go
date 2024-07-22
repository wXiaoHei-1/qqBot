package service

import (
	"context"
	"fmt"
	"log"
	"qqbot/common/types"
	"qqbot/constant"
	"qqbot/utils"
	"runtime"
	"time"
)

// NewSessionManager 获得 session manager 实例
func NewSessionManager() *ChanManager {
	return New()
}

// GetWSS 获取 WebSocket 接入点信息
func (client *HttpClient) GetWSS(ctx context.Context) (*types.WebsocketAP, error) {
	resp, err := client.restyClient.R().SetContext(ctx).
		SetResult(types.WebsocketAP{}).
		Get(constant.GatewayURL)
	if err != nil {
		return nil, err
	}

	return resp.Result().(*types.WebsocketAP), nil
}

// PostMessage 发送消息
func (client *HttpClient) PostMessage(ctx context.Context, channelID string, msg *types.MessageToCreate) (*types.Message, error) {
	resp, err := client.restyClient.R().SetContext(ctx).
		SetResult(types.Message{}).
		SetPathParam("channel_id", channelID).
		SetBody(msg).
		Post(constant.ChannelsURL)
	if err != nil {
		return nil, err
	}

	return resp.Result().(*types.Message), nil
}

// Token 用于调用接口的 token 结构
type Token struct {
	AppID       uint64
	AccessToken string
	Type        string
}

// ToStr 将 Token 转换为字符串形式
func (tk *Token) ToStr() string {
	return fmt.Sprintf("%v.%s", tk.AppID, tk.AccessToken)
}

// New 创建本地 session manager 实例
func New() *ChanManager {
	return &ChanManager{}
}

// ChanManager 默认的本地 session manager 实现
type ChanManager struct {
	sessionChan chan Session
}

// Start 启动本地 session manager
func (l *ChanManager) Start(apInfo *types.WebsocketAP, token *Token, intents int) error {
	// 计算每个 session 的启动间隔时间,避免超过频控限制
	startInterval := utils.CalcInterval(apInfo.SessionStartLimit.MaxConcurrency)
	log.Printf("[ws/session/local] will start %d sessions and per session start interval is %s",
		apInfo.Shards, startInterval)

	// 按照 shards 数量初始化用于启动连接的管理 channel
	l.sessionChan = make(chan Session, apInfo.Shards)
	for i := uint32(0); i < apInfo.Shards; i++ {
		session := Session{
			URL:     apInfo.URL,
			Token:   *token,
			Intent:  intents,
			LastSeq: 0,
			Shards: types.ShardConfig{
				ShardID:    i,
				ShardCount: apInfo.Shards,
			},
		}
		l.sessionChan <- session
	}

	// 启动每个 session 连接
	for session := range l.sessionChan {
		// MaxConcurrency 代表的是每 5s 可以连多少个请求,因此需要控制每个 session 的启动间隔
		time.Sleep(startInterval)
		go l.newConnect(session)
	}
	return nil
}

// PanicHandler 处理websocket场景的 panic ，打印堆栈
func PanicHandler(e interface{}, session *Session) {
	buf := make([]byte, 1024)
	buf = buf[:runtime.Stack(buf, false)]
	log.Printf("[PANIC]%v\n%v\n%s\n", session, e, buf)
}
