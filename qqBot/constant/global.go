package constant

// OpMeans op 对应的含义字符串标识
var OpMeans = map[int]string{
	WSDispatchEvent:  "Event",
	WSHeartbeat:      "Heartbeat",
	WSIdentity:       "Identity",
	WSReTry:          "ReTry",
	WSReconnect:      "Reconnect",
	WSInvalidSession: "InvalidSession",
	WSHello:          "Hello",
	WSHeartbeatAck:   "HeartbeatAck",
}
