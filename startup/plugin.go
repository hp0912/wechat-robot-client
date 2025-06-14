package startup

import (
	"wechat-robot-client/plugin"
	"wechat-robot-client/plugin/plugins"
	"wechat-robot-client/vars"
)

func RegisterMessagePlugin() {
	vars.MessagePlugin = plugin.NewMessagePlugin()
	// 群聊聊天插件
	vars.MessagePlugin.Register(plugins.OnChatRoomAIChatSessionStart)
	vars.MessagePlugin.Register(plugins.OnChatRoomAIChatSessionEnd)
	vars.MessagePlugin.Register(plugins.OnChatRoomAIChat)
	// 群聊绘画插件
	vars.MessagePlugin.Register(plugins.OnChatRoomAIDrawingSessionStart)
	vars.MessagePlugin.Register(plugins.OnChatRoomAIDrawingSessionEnd)
	vars.MessagePlugin.Register(plugins.OnChatRoomAIDrawing)
	// 朋友聊天插件
	vars.MessagePlugin.Register(plugins.OnFriendAIChat)
	// 朋友绘画插件
	vars.MessagePlugin.Register(plugins.OnFriendAIDrawingSessionStart)
	vars.MessagePlugin.Register(plugins.OnFriendAIDrawingSessionEnd)
	vars.MessagePlugin.Register(plugins.OnFriendAIDrawing)
}
