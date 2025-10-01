package plugins

import (
	"context"
	"regexp"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type AutoJoinGroupPlugin struct{}

func NewAutoJoinGroupPlugin() plugin.MessageHandler {
	return &AutoJoinGroupPlugin{}
}

func (p *AutoJoinGroupPlugin) GetName() string {
	return "Auto Join Group"
}

func (p *AutoJoinGroupPlugin) GetLabels() []string {
	return []string{"text", "auto"}
}

func (p *AutoJoinGroupPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AutoJoinGroupPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AutoJoinGroupPlugin) Run(ctx *plugin.MessageContext) bool {
	re := regexp.MustCompile(`^申请进群\s+`)
	chatRoomName := re.ReplaceAllString(ctx.MessageContent, "")
	if chatRoomName == "" {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "群聊名称不能为空")
		return true
	}
	err := service.NewChatRoomService(context.Background()).AutoInviteChatRoomMember(chatRoomName, []string{ctx.Message.FromWxID})
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
	}
	return true
}
