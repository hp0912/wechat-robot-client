package startup

import (
	"wechat-robot-client/plugin"
	"wechat-robot-client/plugin/plugins"
	"wechat-robot-client/vars"
)

func RegisterMessagePlugin() {
	vars.MessagePlugin = plugin.NewMessagePlugin()
	// 群聊聊天插件
	vars.MessagePlugin.Register(plugins.NewChatRoomAIChatPlugin())
	// 朋友聊天插件
	vars.MessagePlugin.Register(plugins.NewFriendAIChatPlugin())
	// 群聊拍一拍交互插件
	vars.MessagePlugin.Register(plugins.NewPatPlugin())
	// 抖音解析插件
	vars.MessagePlugin.Register(plugins.NewDouyinVideoParsePlugin())
	// 图片自动上传插件
	vars.MessagePlugin.Register(plugins.NewImageAutoUploadPlugin())
}
