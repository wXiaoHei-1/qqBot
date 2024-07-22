package service

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"net"
	"net/http"
	"time"
)

type HttpClient struct {
	AppID       uint64
	AccessToken string
	tokenType   string
	timeout     time.Duration
	restyClient *resty.Client // resty client 复用
}

// NewClient 函数: 创建一个新的 HttpClient 实例
func NewClient(ID uint64, token string, duration time.Duration) *HttpClient {
	client := &HttpClient{
		AppID:       ID,
		AccessToken: token,
		timeout:     duration,
		tokenType:   "Bot",
	}

	client.restyClient = resty.New().
		SetTransport(newTransport(nil, 1000)). // 自定义 transport
		SetTimeout(client.timeout).
		SetAuthToken(fmt.Sprintf("%v.%s", client.AppID, client.AccessToken)).
		SetAuthScheme(client.tokenType)

	return client
}

// newTransport 创建一个自定义的 http.Transport 实例
func newTransport(localAddr net.Addr, maxIdleConns int) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   60 * time.Second, // 连接超时时间
		KeepAlive: 60 * time.Second, // TCP KeepAlive 时间
	}
	if localAddr != nil {
		dialer.LocalAddr = localAddr // 设置本地地址
	}

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment, // 使用环境变量中的代理配置
		DialContext:           dialer.DialContext,        // 使用自定义的 dialer
		ForceAttemptHTTP2:     true,                      // 强制使用 HTTP/2
		MaxIdleConns:          maxIdleConns,              // 最大空闲连接数
		IdleConnTimeout:       90 * time.Second,          // 空闲连接最大存活时间
		TLSHandshakeTimeout:   10 * time.Second,          // TLS 握手超时时间
		ExpectContinueTimeout: 1 * time.Second,           // Expect-Continue 超时时间
		MaxIdleConnsPerHost:   maxIdleConns,              // 每个主机的最大空闲连接数
		MaxConnsPerHost:       maxIdleConns,              // 每个主机的最大连接数
	}
}

// newConnect 建立新的 WebSocket 连接,并处理连接成功或失败的情况
func (l *ChanManager) newConnect(session Session) {
	defer func() {
		// panic 留下日志，放回 session
		if err := recover(); err != nil {
			PanicHandler(err, &session)
			l.sessionChan <- session
		}
	}()

	wsClient := NewWebsocket(session)
	if err := wsClient.Connect(); err != nil {
		log.Println(err)
		l.sessionChan <- session // 连接失败，丢回去队列排队重连
		return
	}

	var err error
	if session.ID != "" {
		err = wsClient.ReTry()
	} else {
		err = wsClient.Identify()
	}
	if err != nil {
		log.Printf("[ws/session] Identify/Resume err %+v", err)
		return
	}

	if err = wsClient.Listening(); err != nil {
		log.Printf("[ws/session] Listening err %+v", err)
		currentSession := wsClient.GetSession()
		l.sessionChan <- *currentSession
		return
	}
}
