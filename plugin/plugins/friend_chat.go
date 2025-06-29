package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type FriendAIChatPlugin struct{}

func NewFriendAIChatPlugin() plugin.MessageHandler {
	return &FriendAIChatPlugin{}
}

func (p *FriendAIChatPlugin) GetName() string {
	return "FriendAIChat"
}

func (p *FriendAIChatPlugin) GetLabels() []string {
	return []string{"chat"}
}

func (p *FriendAIChatPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *FriendAIChatPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *FriendAIChatPlugin) Run(ctx *plugin.MessageContext) bool {
	if ctx.Message.IsChatRoom {
		return false
	}
	aiChatService := service.NewAIChatService(ctx.Context, ctx.Settings)
	isAIEnabled := ctx.Settings.IsAIChatEnabled()
	if isAIEnabled {
		defer func() {
			aiChatService.RenewAISession(ctx.Message)
			err := ctx.MessageService.SetMessageIsInContext(ctx.Message)
			if err != nil {
				log.Printf("更新消息上下文失败: %v", err)
			}
		}()
		OnChatIntention(ctx)
		return true
	}
	return false
}
