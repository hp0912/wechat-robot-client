package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type ChatRoomAIChatSessionStartPlugin struct{}

func NewChatRoomAIChatSessionStartPlugin() plugin.MessageHandler {
	return &ChatRoomAIChatSessionStartPlugin{}
}

func (p *ChatRoomAIChatSessionStartPlugin) GetName() string {
	return "ChatRoomAIChatSessionStart"
}

func (p *ChatRoomAIChatSessionStartPlugin) GetLabels() []string {
	return []string{"chat"}
}

func (p *ChatRoomAIChatSessionStartPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *ChatRoomAIChatSessionStartPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *ChatRoomAIChatSessionStartPlugin) Run(ctx *plugin.MessageContext) bool {
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

type ChatRoomAIChatSessionEndPlugin struct{}

func NewChatRoomAIChatSessionEndPlugin() plugin.MessageHandler {
	return &ChatRoomAIChatSessionEndPlugin{}
}

func (p *ChatRoomAIChatSessionEndPlugin) GetName() string {
	return "ChatRoomAIChatSessionEnd"
}

func (p *ChatRoomAIChatSessionEndPlugin) GetLabels() []string {
	return []string{"chat"}
}

func (p *ChatRoomAIChatSessionEndPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *ChatRoomAIChatSessionEndPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *ChatRoomAIChatSessionEndPlugin) Run(ctx *plugin.MessageContext) bool {
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

type ChatRoomAIChatPlugin struct{}

func NewChatRoomAIChatPlugin() plugin.MessageHandler {
	return &ChatRoomAIChatPlugin{}
}

func (p *ChatRoomAIChatPlugin) GetName() string {
	return "ChatRoomAIChat"
}

func (p *ChatRoomAIChatPlugin) GetLabels() []string {
	return []string{"text", "chat"}
}

func (p *ChatRoomAIChatPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *ChatRoomAIChatPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *ChatRoomAIChatPlugin) Run(ctx *plugin.MessageContext) bool {
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
			OnChatIntention(ctx)
			return true
		}
	}
	return false
}
