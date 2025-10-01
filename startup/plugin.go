package startup

import (
	"wechat-robot-client/plugin"
	"wechat-robot-client/plugin/plugins"
	"wechat-robot-client/vars"
)

func RegisterMessagePlugin() {
	vars.MessagePlugin = plugin.NewMessagePlugin()
	// 群聊聊天插件
	vars.MessagePlugin.Register(plugins.NewChatRoomAIChatSessionStartPlugin())
	vars.MessagePlugin.Register(plugins.NewChatRoomAIChatSessionEndPlugin())
	vars.MessagePlugin.Register(plugins.NewChatRoomAIChatPlugin())
	// 群聊绘画插件
	vars.MessagePlugin.Register(plugins.NewChatRoomAIDrawingSessionStartPlugin())
	vars.MessagePlugin.Register(plugins.NewChatRoomAIDrawingSessionEndPlugin())
	vars.MessagePlugin.Register(plugins.NewChatRoomAIDrawingPlugin())
	// 朋友聊天插件
	vars.MessagePlugin.Register(plugins.NewFriendAIChatPlugin())
	// 朋友绘画插件
	vars.MessagePlugin.Register(plugins.NewFriendAIDrawingSessionStartPlugin())
	vars.MessagePlugin.Register(plugins.NewFriendAIDrawingSessionEndPlugin())
	vars.MessagePlugin.Register(plugins.NewFriendAIDrawingPlugin())
	// 群聊拍一拍交互插件
	vars.MessagePlugin.Register(plugins.NewPatPlugin())
	// 图片自动上传插件
	vars.MessagePlugin.Register(plugins.NewImageAutoUploadPlugin())
}
