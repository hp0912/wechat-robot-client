package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

func OnChatRoomAIDrawingSessionStart(ctx *plugin.MessageContext) bool {
	if !ctx.Message.IsChatRoom {
		return false
	}
	aiChatService := service.NewAIChatService(ctx.Context, ctx.Settings)
	aiDrawingService := service.NewAIDrawingService(ctx.Context, ctx.Settings)
	if aiDrawingService.IsAISessionStart(ctx.Message) {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiDrawingService.GetAISessionStartTips(), ctx.Message.SenderWxID)
		// 是绘画，则结束闲聊上下文，如果有的话
		err := aiChatService.ExpireAISession(ctx.Message)
		if err != nil {
			log.Println("结束闲聊会话失败:", err)
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

func OnChatRoomAIDrawingSessionEnd(ctx *plugin.MessageContext) bool {
	if !ctx.Message.IsChatRoom {
		return false
	}
	aiDrawingService := service.NewAIDrawingService(ctx.Context, ctx.Settings)
	if aiDrawingService.IsAISessionEnd(ctx.Message) {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiDrawingService.GetAISessionEndTips(), ctx.Message.SenderWxID)
		return true
	}
	return false
}

func OnChatRoomAIDrawing(ctx *plugin.MessageContext) bool {
	if !ctx.Message.IsChatRoom {
		return false
	}
	aiDrawingService := service.NewAIDrawingService(ctx.Context, ctx.Settings)
	isInSession, err := aiDrawingService.IsInAISession(ctx.Message)
	if err != nil {
		log.Printf("检查AI绘图会话失败: %v", err)
		return true
	}
	isAIEnabled := ctx.Settings.IsAIDrawingEnabled()
	if isAIEnabled {
		if isInSession {
			defer func() {
				aiDrawingService.RenewAISession(ctx.Message)
				err := ctx.MessageService.SetMessageIsInContext(ctx.Message)
				if err != nil {
					log.Printf("更新消息上下文失败: %v", err)
				}
			}()
			OnAIDrawing(ctx)
			return true
		}
	}
	return false
}
