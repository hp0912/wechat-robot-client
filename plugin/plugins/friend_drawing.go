package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type FriendAIDrawingSessionStartPlugin struct{}

func NewFriendAIDrawingSessionStartPlugin() plugin.MessageHandler {
	return &FriendAIDrawingSessionStartPlugin{}
}

func (p *FriendAIDrawingSessionStartPlugin) GetName() string {
	return "FriendAIDrawingSessionStart"
}

func (p *FriendAIDrawingSessionStartPlugin) GetLabels() []string {
	return []string{"text", "drawing"}
}

func (p *FriendAIDrawingSessionStartPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *FriendAIDrawingSessionStartPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *FriendAIDrawingSessionStartPlugin) Run(ctx *plugin.MessageContext) bool {
	if ctx.Message.IsChatRoom {
		return false
	}
	aiDrawingService := service.NewAIDrawingService(ctx.Context, ctx.Settings)
	if aiDrawingService.IsAISessionStart(ctx.Message) {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiDrawingService.GetAISessionStartTips(), ctx.Message.SenderWxID)
		// 重置一下会话上下文
		err := ctx.MessageService.ResetChatRoomAIMessageContext(ctx.Message)
		if err != nil {
			log.Println("重置会话上下文失败:", err)
		}
		return true
	}
	return false
}

type FriendAIDrawingSessionEndPlugin struct{}

func NewFriendAIDrawingSessionEndPlugin() plugin.MessageHandler {
	return &FriendAIDrawingSessionEndPlugin{}
}

func (p *FriendAIDrawingSessionEndPlugin) GetName() string {
	return "FriendAIDrawingSessionEnd"
}

func (p *FriendAIDrawingSessionEndPlugin) GetLabels() []string {
	return []string{"drawing"}
}

func (p *FriendAIDrawingSessionEndPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *FriendAIDrawingSessionEndPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *FriendAIDrawingSessionEndPlugin) Run(ctx *plugin.MessageContext) bool {
	if ctx.Message.IsChatRoom {
		return false
	}
	aiDrawingService := service.NewAIDrawingService(ctx.Context, ctx.Settings)
	if aiDrawingService.IsAISessionEnd(ctx.Message) {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiDrawingService.GetAISessionEndTips(), ctx.Message.SenderWxID)
		return true
	}
	return false
}

type FriendAIDrawingPlugin struct{}

func NewFriendAIDrawingPlugin() plugin.MessageHandler {
	return &FriendAIDrawingPlugin{}
}

func (p *FriendAIDrawingPlugin) GetName() string {
	return "FriendAIDrawing"
}

func (p *FriendAIDrawingPlugin) GetLabels() []string {
	return []string{"drawing"}
}

func (p *FriendAIDrawingPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *FriendAIDrawingPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *FriendAIDrawingPlugin) Run(ctx *plugin.MessageContext) bool {
	if ctx.Message.IsChatRoom {
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
