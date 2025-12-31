package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
)

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
	return NewChatRoomCommonPlugin().PreAction(ctx)
}

func (p *ChatRoomAIChatPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *ChatRoomAIChatPlugin) Run(ctx *plugin.MessageContext) bool {
	if !p.PreAction(ctx) {
		return false
	}
	isAIEnabled := ctx.Settings.IsAIChatEnabled()
	isAITrigger := ctx.Settings.IsAITrigger()
	if isAIEnabled {
		if isAITrigger {
			defer func() {
				err := ctx.MessageService.SetMessageIsInContext(ctx.Message)
				if err != nil {
					log.Printf("更新消息上下文失败: %v", err)
				}
			}()
			aiChat := NewAIChatPlugin()
			aiChat.Run(ctx)
			return true
		}
	}
	return false
}
