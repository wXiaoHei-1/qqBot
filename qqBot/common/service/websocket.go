package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"qqbot/common/types"
	constant "qqbot/constant"
	"qqbot/utils"
	"syscall"
	"time"

	wss "github.com/gorilla/websocket"
)

var ResumeSignal syscall.Signal

//type messageChan chan *types.WSPayload
//type closeErrorChan chan error

// Session 连接的 session 结构，包括链接的所有必要字段
type Session struct {
	ID      string
	URL     string
	Token   Token
	Intent  int
	LastSeq uint32
	Shards  types.ShardConfig
}

// WebsocketClient Client websocket 连接客户端
type WebsocketClient struct {
	Version         int
	Conn            *wss.Conn
	MessageQueue    types.MessageChan
	Session         *Session
	User            *types.WSUser
	CloseChan       types.CloseErrorChan
	HeartBeatTicker *time.Ticker // 用于维持定时心跳
}

// NewWebsocket 创建一个新的 ws 实例，需要传递 session 对象
func NewWebsocket(session Session) *WebsocketClient {
	return &WebsocketClient{
		MessageQueue:    make(types.MessageChan, 2000),
		Session:         &session,
		CloseChan:       make(types.CloseErrorChan, 10),
		HeartBeatTicker: time.NewTicker(60 * time.Second), // 先给一个默认 ticker，在收到 hello 包之后，会 reset
	}
}

// Connect 连接到 wss 地址
func (c *WebsocketClient) Connect() error {
	if c.Session.URL == "" {
		return errors.New("websocket url is invalid")
	}

	var err error
	c.Conn, _, err = wss.DefaultDialer.Dial(c.Session.URL, nil)
	if err != nil {
		log.Printf("%s, connect err: %v", c.Session, err)
		return err
	}
	log.Printf("%s, url %s, connected", c.Session, c.Session.URL)
	return nil
}

// Identify 鉴权
func (c *WebsocketClient) Identify() error {
	// 避免传错 intent
	if c.Session.Intent == 0 {
		c.Session.Intent = 1
	}
	payload := &types.WSPayload{
		Data: &types.WSIdentityData{
			Token:   c.Session.Token.ToStr(),
			Intents: c.Session.Intent,
			Shard: []uint32{
				c.Session.Shards.ShardID,
				c.Session.Shards.ShardCount,
			},
		},
	}
	payload.OPCode = constant.WSIdentity
	return c.SendMessage(payload)
}

// GetSession 拉取 session 信息，包括 token，shard，seq 等
func (c *WebsocketClient) GetSession() *Session {
	return c.Session
}

// ReTry 重连
func (c *WebsocketClient) ReTry() error {
	payload := &types.WSPayload{
		Data: &types.WSResumeData{
			Token:     c.Session.Token.ToStr(),
			SessionID: c.Session.ID,
			Seq:       c.Session.LastSeq,
		},
	}
	payload.OPCode = constant.WSReTry // 内嵌结构体字段，单独赋值
	return c.SendMessage(payload)
}

// Listening 监听 websocket 事件
func (c *WebsocketClient) Listening() error {
	defer c.Close()
	// 读取消息到队列
	go c.readMessageToQueue()
	// 从队列读取消息并处理，在 goroutine 中执行以避免业务逻辑阻塞 closeChan 和 heartBeatTicker
	go c.listenMessageAndHandle()

	// 接收重连信号
	resumeSignal := make(chan os.Signal, 1)
	if ResumeSignal >= syscall.SIGHUP {
		signal.Notify(resumeSignal, ResumeSignal)
	}

	// 处理消息
	for {
		select {
		case <-resumeSignal: // 使用信号量控制连接立即重连
			log.Printf("%s, received resumeSignal signal", c.Session)
			return errors.New("need reconnect")
		case err := <-c.CloseChan:
			log.Printf("%v Listening stop. err is %v", c.Session, err)
			if wss.IsCloseError(err, 4914, 4915) {
				err = errors.New(err.Error())
			}
			if wss.IsUnexpectedCloseError(err, 4009) {
				err = errors.New(err.Error())
			}
			return err
		case <-c.HeartBeatTicker.C:
			log.Printf("%v listened heartBeat", c.Session)
			heartBeatEvent := &types.WSPayload{
				WSPayloadBase: types.WSPayloadBase{
					OPCode: constant.WSHeartbeat,
				},
				Data: c.Session.LastSeq,
			}
			// 不处理错误，Write 内部会处理，如果发生发包异常，会通知主协程退出
			_ = c.SendMessage(heartBeatEvent)
		}
	}
}

// SendMessage 发送数据
func (c *WebsocketClient) SendMessage(message *types.WSPayload) error {
	m, _ := json.Marshal(message)
	log.Printf("%v write %s message, %v", c.Session, utils.GetOpMeans(message.OPCode), string(m))

	if err := c.Conn.WriteMessage(wss.TextMessage, m); err != nil {
		log.Printf("%s WriteMessage failed, %v", c.Session, err)
		c.CloseChan <- err
		return err
	}
	return nil
}

// Close 关闭连接
func (c *WebsocketClient) Close() {
	if err := c.Conn.Close(); err != nil {
		log.Printf("%s, close conn err: %v", c.Session, err)
	}
	c.HeartBeatTicker.Stop()
}

// readMessageToQueue 从 WebSocket 连接中读取消息,解析并投递到消息队列
func (c *WebsocketClient) readMessageToQueue() {
	for {
		// 从 WebSocket 连接中读取消息
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			// 读取消息失败,打印错误日志,关闭消息队列,并通知关闭连接
			log.Printf("%s read message failed, %v, message %s", c.Session, err, string(message))
			close(c.MessageQueue)
			c.CloseChan <- err
			return
		}

		// 解析消息为 WSPayload 结构
		payload := &types.WSPayload{}
		if err := json.Unmarshal(message, payload); err != nil {
			// 消息解析失败,打印错误日志并继续下一个消息
			log.Printf("%s json failed, %v", c.Session, err)
			continue
		}
		payload.RawMessage = message
		log.Printf("%s receive %s message, %s", c.Session, utils.GetOpMeans(payload.OPCode), string(message))

		// 处理内置的一些事件,如果处理成功,则不再投递给业务
		if c.isHandleBuildIn(payload) {
			continue
		}
		// 将 WSPayload 投递到消息队列中
		c.MessageQueue <- payload
	}
}

// listenMessageAndHandle WebSocket 消息队列中读取事件,根据事件类型进行相应的处理,包括捕获可能发生的异常并进行重连
func (c *WebsocketClient) listenMessageAndHandle() {
	defer func() {
		// panic，一般是由于业务自己实现的 handle 不完善导致
		// 打印日志后，关闭这个连接，进入重连流程
		if err := recover(); err != nil {
			PanicHandler(err, c.Session)
			c.CloseChan <- fmt.Errorf("panic: %v", err)
		}
	}()
	for payload := range c.MessageQueue {
		if payload.Seq > 0 {
			c.Session.LastSeq = payload.Seq
		}
		// ready 事件需要特殊处理
		if payload.Type == "READY" {
			c.readyHandler(payload)
			continue
		}
		// 解析具体事件，并投递给业务注册的 handler
		if err := ParseAndHandle(payload); err != nil {
			log.Printf("%s parseAndHandle failed, %v", c.Session, err)
		}
	}
	log.Printf("%s message queue is closed", c.Session)
}

// isHandleBuildIn 内置的事件处理，处理那些不需要业务方处理的事件
// return true 的时候说明事件已经被处理了
func (c *WebsocketClient) isHandleBuildIn(payload *types.WSPayload) bool {
	switch payload.OPCode {
	case constant.WSHello: // 接收到 hello 后需要开始发心跳
		c.startHeartBeatTicker(payload.RawMessage)
	case constant.WSHeartbeatAck: // 心跳 ack 不需要业务处理
	case constant.WSReconnect: // 达到连接时长，需要重新连接，此时可以通过 resume 续传原连接上的事件
		c.CloseChan <- errors.New("need reconnect")
	case constant.WSInvalidSession: // 无效的 sessionLog，需要重新鉴权
		c.CloseChan <- errors.New("invalid session")
	default:
		return false
	}
	return true
}

// startHeartBeatTicker 启动定时心跳
func (c *WebsocketClient) startHeartBeatTicker(message []byte) {
	helloData := &types.WSHelloData{}
	if err := utils.ParseData(message, helloData); err != nil {
		log.Printf("%s hello data parse failed, %v, message %v", c.Session, err, message)
	}
	// 根据 hello 的回包，重新设置心跳的定时器时间
	c.HeartBeatTicker.Reset(time.Duration(helloData.HeartbeatInterval) * time.Millisecond)
}

// readyHandler 针对ready返回的处理，需要记录 sessionID 等相关信息
func (c *WebsocketClient) readyHandler(payload *types.WSPayload) {
	readyData := &types.WSReadyData{}
	if err := utils.ParseData(payload.RawMessage, readyData); err != nil {
		log.Printf("%v parseReadyData failed, %v, message %v", c.Session, err, payload.RawMessage)
	}
	c.Version = readyData.Version
	// 基于 ready 事件，更新 session 信息
	c.Session.ID = readyData.SessionID
	c.Session.Shards.ShardID = readyData.Shard[0]
	c.Session.Shards.ShardCount = readyData.Shard[1]
	c.User = &types.WSUser{
		ID:       readyData.User.ID,
		Username: readyData.User.Username,
		Bot:      readyData.User.Bot,
	}
}
