package constant

const (
	GatewayURL             = "https://api.sgroup.qq.com/gateway/bot"
	ChannelsURL            = "https://api.sgroup.qq.com/channels/{channel_id}/messages"
	DashScopeAPIURL string = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	DashScopeModel  string = "qwen-turbo"
)

// WS OPCode
const (
	WSDispatchEvent int = iota
	WSHeartbeat
	WSIdentity
	_ // Presence Update
	_ // Voice State Update
	_
	WSReTry
	WSReconnect
	_ // Request Guild Members
	WSInvalidSession
	WSHello
	WSHeartbeatAck
	HTTPCallbackAck
)
