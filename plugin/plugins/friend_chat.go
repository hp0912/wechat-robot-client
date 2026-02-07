package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/vars"
)

type FriendAIChatPlugin struct{}

func NewFriendAIChatPlugin() plugin.MessageHandler {
	return &FriendAIChatPlugin{}
}

func (p *FriendAIChatPlugin) GetName() string {
	return "FriendAIChat"
}

func (p *FriendAIChatPlugin) GetLabels() []string {
	return []string{"text", "chat"}
}

func (p *FriendAIChatPlugin) Match(ctx *plugin.MessageContext) bool {
	return !ctx.Message.IsChatRoom
}

func (p *FriendAIChatPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *FriendAIChatPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *FriendAIChatPlugin) Run(ctx *plugin.MessageContext) {
	// 修复 AI 会响应自己发送(从其他设备)的消息的问题
	if ctx.Message != nil && ctx.Message.SenderWxID == vars.RobotRuntime.WxID {
		return
	}
	isAIEnabled := ctx.Settings.IsAIChatEnabled()
	if isAIEnabled {
		defer func() {
			err := ctx.MessageService.SetMessageIsInContext(ctx.Message)
			if err != nil {
				log.Printf("更新消息上下文失败: %v", err)
			}
		}()
		aiChat := NewAIChatPlugin()
		if !aiChat.Match(ctx) {
			return
		}
		aiChat.Run(ctx)
		return
	}
}
