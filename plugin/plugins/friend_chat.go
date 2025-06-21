package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

func OnFriendAIChat(ctx *plugin.MessageContext) bool {
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
