package service

import (
	"qqbot/common/types"
	constant "qqbot/constant"
	"qqbot/utils"
)

// ATMessageEventHandler 处理 AT 消息事件的回调函数
type ATMessageEventHandler func(event *types.WSPayload, data *types.Message) error

// eventParseFunc 解析 WebSocket 事件的回调函数
type eventParseFunc func(event *types.WSPayload, message []byte) error

// DefaultHandlers 管理所有支持的事件处理器
var DefaultHandlers struct {
	ATMessage ATMessageEventHandler
}

// RegisterHandlers 注册事件回调,并返回用于 WebSocket 鉴权的 intent
func RegisterHandlers(handlers ...interface{}) int {
	var intent int
	for _, h := range handlers {
		switch handle := h.(type) {
		case ATMessageEventHandler:
			DefaultHandlers.ATMessage = handle
			intent |= 1 << 30 //使用位运算 |= 修改 intent 变量,设置相应的位为 1 来表示已注册该事件类型
		}
	}
	return intent
}

// ParseAndHandle 处理回调事件
func ParseAndHandle(payload *types.WSPayload) error {
	// 根据 opcode 和事件类型查找对应的处理函数
	if parseFunc, ok := eventParseFuncMap[payload.OPCode][payload.Type]; ok {
		return parseFunc(payload, payload.RawMessage)
	}
	return nil
}

// eventParseFunc 解析 WebSocket 事件的回调函数
var eventParseFuncMap = map[int]map[string]eventParseFunc{
	constant.WSDispatchEvent: {
		"AT_MESSAGE_CREATE": atMessageHandler,
	},
}

// atMessageHandler 解析 AT 消息事件的数据,并调用注册的 AT 消息事件处理器
func atMessageHandler(payload *types.WSPayload, message []byte) error {
	// 1. 解析 AT 消息事件的数据
	data := &types.Message{}
	if err := utils.ParseData(message, data); err != nil {
		return err
	}
	// 2. 检查是否已注册 AT 消息事件的处理器
	if DefaultHandlers.ATMessage == nil {
		return nil // 如果没有注册处理器,则什么也不做
	}
	// 3. 调用注册的 AT 消息事件处理器
	return DefaultHandlers.ATMessage(payload, data)
}
