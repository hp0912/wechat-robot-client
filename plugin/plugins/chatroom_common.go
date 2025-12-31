package plugins

import (
	"log"
	"wechat-robot-client/interface/plugin"
)

type ChatRoomCommonPlugin struct{}

func NewChatRoomCommonPlugin() plugin.MessageHandler {
	return &ChatRoomCommonPlugin{}
}

func (p *ChatRoomCommonPlugin) GetName() string {
	return "ChatRoomCommon"
}

func (p *ChatRoomCommonPlugin) GetLabels() []string {
	return []string{"text", "chat"}
}

func (p *ChatRoomCommonPlugin) PreAction(ctx *plugin.MessageContext) bool {
	if !ctx.Message.IsChatRoom {
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
	return true
}

func (p *ChatRoomCommonPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *ChatRoomCommonPlugin) Run(ctx *plugin.MessageContext) bool {
	return true
}
