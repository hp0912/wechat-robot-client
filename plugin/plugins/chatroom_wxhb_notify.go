package plugins

import (
	"context"
	"log"
	"strings"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type ChatRoomWxhbNotifyPlugin struct {
	NotifyMemberList []string
}

func NewChatRoomWxhbNotifyPlugin() plugin.MessageHandler {
	return &ChatRoomWxhbNotifyPlugin{}
}

func (p *ChatRoomWxhbNotifyPlugin) GetName() string {
	return "ChatRoomWxhbNotify"
}

func (p *ChatRoomWxhbNotifyPlugin) GetLabels() []string {
	return []string{"red-envelopes", "chat"}
}

func (p *ChatRoomWxhbNotifyPlugin) Match(ctx *plugin.MessageContext) bool {
	return ctx.Message.Type == model.MsgTypeApp && ctx.Message.AppMsgType == model.AppMsgTypeRedEnvelopes
}

func (p *ChatRoomWxhbNotifyPlugin) PreAction(ctx *plugin.MessageContext) bool {
	if !NewChatRoomCommonPlugin().PreAction(ctx) {
		return false
	}
	chatRoomSettings, err := service.NewChatRoomSettingsService(context.Background()).GetChatRoomSettings(ctx.Message.FromWxID)
	if err != nil {
		log.Printf("获取群设置失败: %v", err)
		return false
	}
	if chatRoomSettings == nil {
		log.Printf("群设置不存在: 群ID=%s", ctx.Message.FromWxID)
		return false
	}
	if chatRoomSettings.WxhbNotifyEnabled == nil || !*chatRoomSettings.WxhbNotifyEnabled {
		log.Printf("群红包通知未开启: 群ID=%s", ctx.Message.FromWxID)
		return false
	}
	if chatRoomSettings.WxhbNotifyMemberList == nil || *chatRoomSettings.WxhbNotifyMemberList == "" {
		log.Printf("群红包通知成员列表为空: 群ID=%s", ctx.Message.FromWxID)
		return false
	}
	p.NotifyMemberList = strings.Split(*chatRoomSettings.WxhbNotifyMemberList, ",")
	return true
}

func (p *ChatRoomWxhbNotifyPlugin) PostAction(ctx *plugin.MessageContext) {
}

func (p *ChatRoomWxhbNotifyPlugin) Run(ctx *plugin.MessageContext) {
	if !p.PreAction(ctx) {
		return
	}

	var xmlMessage robot.XmlMessage
	err := vars.RobotRuntime.XmlDecoder(ctx.Message.Content, &xmlMessage)
	if err != nil {
		log.Printf("解析红包消息XML失败: %v", err)
		return
	}

	if xmlMessage.AppMsg.WcPayInfo.SceneID == "1001" {
		log.Println("群收款通知~")
		return
	}

	notifyTargets := p.buildNotifyTargets(ctx.Message.SenderWxID, xmlMessage.AppMsg.WcPayInfo.ExclusiveRecvUsername)
	if len(notifyTargets) == 0 {
		return
	}

	_ = ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "红包来啦~", notifyTargets...)
}

func (p *ChatRoomWxhbNotifyPlugin) buildNotifyTargets(senderWxID, exclusiveRecvUsername string) []string {
	senderWxID = strings.TrimSpace(senderWxID)
	exclusiveRecvUsername = strings.TrimSpace(exclusiveRecvUsername)

	uniqueNotifyMembers := make([]string, 0, len(p.NotifyMemberList))
	memberSet := make(map[string]struct{}, len(p.NotifyMemberList))
	for _, member := range p.NotifyMemberList {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		if member == senderWxID {
			continue
		}
		if _, exists := memberSet[member]; exists {
			continue
		}
		memberSet[member] = struct{}{}
		uniqueNotifyMembers = append(uniqueNotifyMembers, member)
	}

	if exclusiveRecvUsername != "" {
		if _, ok := memberSet[exclusiveRecvUsername]; !ok {
			return []string{}
		}
		return []string{exclusiveRecvUsername}
	}

	return uniqueNotifyMembers
}
