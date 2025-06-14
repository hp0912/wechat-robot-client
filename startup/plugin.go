package startup

import (
	"wechat-robot-client/plugin"
	"wechat-robot-client/plugin/plugins"
	"wechat-robot-client/vars"
)

func RegisterMessagePlugin() {
	vars.MessagePlugin = plugin.NewMessagePlugin()
	// 群聊插件
	vars.MessagePlugin.Register(plugins.OnChatRoomAIChatSessionStart)
	vars.MessagePlugin.Register(plugins.OnChatRoomAIChatSessionEnd)
	vars.MessagePlugin.Register(plugins.OnChatRoomAIChat)
	// 朋友聊天插件
	vars.MessagePlugin.Register(plugins.OnFriendAIChat)
}
