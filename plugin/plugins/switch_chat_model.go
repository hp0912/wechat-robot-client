package plugins

import (
	"context"
	"log"
	"strings"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type SwitchChatModelPlugin struct{}

func NewSwitchChatModelPlugin() plugin.MessageHandler {
	return &SwitchChatModelPlugin{}
}

func (p *SwitchChatModelPlugin) GetName() string {
	return "SwitchChatModel"
}

func (p *SwitchChatModelPlugin) GetLabels() []string {
	return []string{"text", "chat"}
}

func (p *SwitchChatModelPlugin) PreAction(ctx *plugin.MessageContext) bool {
	if ctx.MessageContent == "" || !strings.HasPrefix(ctx.MessageContent, "#切换聊天模型") {
		return false
	}
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

func (p *SwitchChatModelPlugin) PostAction(ctx *plugin.MessageContext) {
}

func (p *SwitchChatModelPlugin) Run(ctx *plugin.MessageContext) bool {
	if !p.PreAction(ctx) {
		return false
	}

	parts := strings.Fields(ctx.MessageContent)
	if len(parts) < 2 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "请提供要切换的聊天模型名称，例如：#切换聊天模型 gpt-3.5-turbo", ctx.Message.SenderWxID)
		return true
	}
	newModel := parts[1]
	if ctx.Message.IsChatRoom {
		svc := service.NewChatRoomSettingsService(context.Background())
		chatRoomSettings, err := svc.GetChatRoomSettings(ctx.Message.FromWxID)
		if err != nil {
			log.Printf("获取群设置失败: %v", err)
			return true
		}
		if chatRoomSettings == nil {
			log.Printf("群设置不存在: 群ID=%s", ctx.Message.FromWxID)
			return true
		}
		chatRoomSettings.ChatModel = &newModel
		err = svc.SaveChatRoomSettings(chatRoomSettings)
		if err != nil {
			log.Printf("保存群设置失败: %v", err)
			return true
		}
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "已将聊天模型切换为："+newModel, ctx.Message.SenderWxID)
	} else {
		friendSettings, err := service.NewFriendSettingsService(context.Background()).GetFriendSettings(ctx.Message.FromWxID)
		if err != nil {
			log.Printf("获取群设置失败: %v", err)
			return true
		}
		if friendSettings == nil {
			log.Printf("群设置不存在: 群ID=%s", ctx.Message.FromWxID)
			return true
		}
		friendSettings.ChatModel = &newModel
		err = service.NewFriendSettingsService(context.Background()).SaveFriendSettings(friendSettings)
		if err != nil {
			log.Printf("保存好友设置失败: %v", err)
			return true
		}
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "已将聊天模型切换为："+newModel, ctx.Message.SenderWxID)
	}
	return true
}
