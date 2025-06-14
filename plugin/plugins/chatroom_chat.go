package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

func OnChatRoomAIChatSessionStart(ctx *plugin.MessageContext) bool {
	if !ctx.Message.IsChatRoom {
		return false
	}
	aiChatService := service.NewAIChatService(ctx.Context, ctx.Settings)
	aiDrawingService := service.NewAIDrawingService(ctx.Context, ctx.Settings)
	if aiChatService.IsAISessionStart(ctx.Message) {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiChatService.GetAISessionStartTips(), ctx.Message.SenderWxID)
		// 是闲聊，则结束绘画上下文，如果有的话
		err := aiDrawingService.ExpireAISession(ctx.Message)
		if err != nil {
			log.Println("结束绘画会话失败:", err)
		}
		// 重置一下会话上下文
		err = ctx.MessageService.ResetChatRoomAIMessageContext(ctx.Message)
		if err != nil {
			log.Println("重置会话上下文失败:", err)
		}
		return true
	}
	return false
}

func OnChatRoomAIChatSessionEnd(ctx *plugin.MessageContext) bool {
	if !ctx.Message.IsChatRoom {
		return false
	}
	aiChatService := service.NewAIChatService(ctx.Context, ctx.Settings)
	if aiChatService.IsAISessionEnd(ctx.Message) {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiChatService.GetAISessionEndTips(), ctx.Message.SenderWxID)
		return true
	}
	return false
}

func OnChatRoomAIChat(ctx *plugin.MessageContext) bool {
	if !ctx.Message.IsChatRoom {
		return false
	}
	aiChatService := service.NewAIChatService(ctx.Context, ctx.Settings)
	isInSession, err := aiChatService.IsInAISession(ctx.Message)
	if err != nil {
		log.Printf("检查AI会话失败: %v", err)
		return true
	}
	isAIEnabled := ctx.Settings.IsAIChatEnabled()
	isAITrigger := ctx.Settings.IsAITrigger()
	if isAIEnabled {
		if isAITrigger || isInSession {
			defer func() {
				aiChatService.RenewAISession(ctx.Message)
				err := ctx.MessageService.SetMessageIsInContext(ctx.Message)
				if err != nil {
					log.Printf("更新消息上下文失败: %v", err)
				}
			}()
			OnChatIntention(ctx, aiChatService)
			return true
		}
	}
	return false
}
