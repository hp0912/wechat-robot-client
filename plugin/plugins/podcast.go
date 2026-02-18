package plugins

import (
	"strings"
	"wechat-robot-client/interface/plugin"
)

type PodcastPlugin struct{}

func NewPodcastPlugin() plugin.MessageHandler {
	return &PodcastPlugin{}
}

func (p *PodcastPlugin) GetName() string {
	return "Podcast"
}

func (p *PodcastPlugin) GetLabels() []string {
	return []string{"text", "chat"}
}

func (p *PodcastPlugin) Match(ctx *plugin.MessageContext) bool {
	return strings.HasPrefix(ctx.MessageContent, "#AI播客")
}

func (p *PodcastPlugin) PreAction(ctx *plugin.MessageContext) bool {
	if !NewChatRoomCommonPlugin().PreAction(ctx) {
		return false
	}
	return true
}

func (p *PodcastPlugin) PostAction(ctx *plugin.MessageContext) {
}

func (p *PodcastPlugin) Run(ctx *plugin.MessageContext) {
	if !p.PreAction(ctx) {
		return
	}
	ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "AI 播客开发中，敬请期待~")
}
