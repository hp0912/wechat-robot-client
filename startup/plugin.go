package startup

import (
	"wechat-robot-client/plugin"
	"wechat-robot-client/vars"
)

func RegisterPlugin() {
	// 定义一个处理器
	dispatcher := plugin.NewMessageMatchDispatcher()
	// 设置为异步处理
	dispatcher.SetAsync(true)

	vars.MessageHandler = plugin.DispatchMessage(dispatcher)
}
