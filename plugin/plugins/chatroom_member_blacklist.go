package plugins

import (
	"context"
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/service"
)

type ChatRoomMemberBlacklistPlugin struct{}

func NewChatRoomMemberBlacklistPlugin() plugin.MessageHandler {
	return &ChatRoomMemberBlacklistPlugin{}
}

func (p *ChatRoomMemberBlacklistPlugin) GetName() string {
	return "ChatRoomMemberBlacklist"
}

func (p *ChatRoomMemberBlacklistPlugin) GetLabels() []string {
	return []string{"text", "chat"}
}

func (p *ChatRoomMemberBlacklistPlugin) Match(ctx *plugin.MessageContext) bool {
	return ctx.Message.IsChatRoom && ctx.ReferMessage != nil && ctx.MessageContent == "#加入黑名单"
}

func (p *ChatRoomMemberBlacklistPlugin) PreAction(ctx *plugin.MessageContext) bool {
	chatRoomMember, err := ctx.MessageService.GetChatRoomMember(ctx.Message.FromWxID, ctx.Message.SenderWxID)
	if err != nil {
		log.Printf("获取群成员信息失败: %v", err)
		return false
	}
	if chatRoomMember == nil {
		log.Printf("群成员信息不存在: 群ID=%s, 成员微信ID=%s", ctx.Message.FromWxID, ctx.Message.SenderWxID)
		return false
	}
	if chatRoomMember.IsBlacklisted != nil && *chatRoomMember.IsBlacklisted {
		log.Printf("群成员[%s]在黑名单中，跳过AI回复", chatRoomMember.Nickname)
		return false
	}
	if chatRoomMember.IsAdmin == nil || !*chatRoomMember.IsAdmin {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "您配使用这个指令吗？", ctx.Message.SenderWxID)
		return false
	}
	return true
}

func (p *ChatRoomMemberBlacklistPlugin) PostAction(ctx *plugin.MessageContext) {
}

func (p *ChatRoomMemberBlacklistPlugin) Run(ctx *plugin.MessageContext) {
	if !p.PreAction(ctx) {
		return
	}
	isBlacklisted := true
	err := service.NewChatRoomService(context.Background()).BatchUpdateChatRoomMemberInfo(model.UpdateChatRoomMember{
		ChatRoomID:    ctx.Message.FromWxID,
		WechatID:      ctx.ReferMessage.SenderWxID,
		IsBlacklisted: &isBlacklisted,
	})
	if err != nil {
		log.Printf("将群成员加入黑名单失败: %v", err)
		return
	}
}
