package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type ChatRoomAIDrawingSessionStartPlugin struct{}

func NewChatRoomAIDrawingSessionStartPlugin() plugin.MessageHandler {
	return &ChatRoomAIDrawingSessionStartPlugin{}
}

func (p *ChatRoomAIDrawingSessionStartPlugin) GetName() string {
	return "ChatRoomAIDrawingSessionStart"
}

func (p *ChatRoomAIDrawingSessionStartPlugin) GetLabels() []string {
	return []string{"text", "drawing"}
}

func (p *ChatRoomAIDrawingSessionStartPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *ChatRoomAIDrawingSessionStartPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *ChatRoomAIDrawingSessionStartPlugin) Run(ctx *plugin.MessageContext) bool {
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

type ChatRoomAIDrawingSessionEndPlugin struct{}

func NewChatRoomAIDrawingSessionEndPlugin() plugin.MessageHandler {
	return &ChatRoomAIDrawingSessionEndPlugin{}
}

func (p *ChatRoomAIDrawingSessionEndPlugin) GetName() string {
	return "ChatRoomAIDrawingSessionEnd"
}

func (p *ChatRoomAIDrawingSessionEndPlugin) GetLabels() []string {
	return []string{"drawing"}
}

func (p *ChatRoomAIDrawingSessionEndPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *ChatRoomAIDrawingSessionEndPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *ChatRoomAIDrawingSessionEndPlugin) Run(ctx *plugin.MessageContext) bool {
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

type ChatRoomAIDrawingPlugin struct{}

func NewChatRoomAIDrawingPlugin() plugin.MessageHandler {
	return &ChatRoomAIDrawingPlugin{}
}

func (p *ChatRoomAIDrawingPlugin) GetName() string {
	return "ChatRoomAIDrawing"
}

func (p *ChatRoomAIDrawingPlugin) GetLabels() []string {
	return []string{"drawing"}
}

func (p *ChatRoomAIDrawingPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *ChatRoomAIDrawingPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *ChatRoomAIDrawingPlugin) Run(ctx *plugin.MessageContext) bool {
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
			aiDrawing := NewAIDrawingPlugin()
			aiDrawing.Run(ctx)
			return true
		}
	}
	return false
}
