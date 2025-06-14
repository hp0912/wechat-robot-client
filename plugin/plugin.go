package plugin

import "wechat-robot-client/interface/plugin"

type MessagePlugin struct {
	Plugins []plugin.MessageHandler
}

func NewMessagePlugin() *MessagePlugin {
	return &MessagePlugin{}
}

func (mp *MessagePlugin) Register(handler plugin.MessageHandler) {
	mp.Plugins = append(mp.Plugins, handler)
}
