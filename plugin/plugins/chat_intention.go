package plugins

import (
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

func OnChatIntention(ctx *plugin.MessageContext, aiChatService *service.AIChatService) {
	aiWorkflowService := service.NewAIWorkflowService(ctx.Context, ctx.Settings)
	chatIntention := aiWorkflowService.ChatIntention(ctx.Message)
	switch chatIntention {
	case service.ChatIntentionChat:
		aiContext, err := ctx.MessageService.GetAIMessageContext(ctx.Message)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return
		}
		aiReply, err := aiChatService.Chat(aiContext)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return
		}
		if ctx.Message.IsChatRoom {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReply, ctx.Message.SenderWxID)
		} else {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReply)
		}
	case service.ChatIntentionSing:
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "唱歌功能正在开发中，敬请期待！")
	case service.ChatIntentionSongRequest:
		title := aiWorkflowService.GetSongRequestTitle(ctx.Message)
		if title == "" {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "抱歉，我无法识别您想要点的歌曲。")
			return
		}
		err := ctx.MessageService.SendMusicMessage(ctx.Message.FromWxID, title)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		}
	case service.ChatIntentionDrawAPicture:
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "绘画功能正在开发中，敬请期待！")
	case service.ChatIntentionEditPictures:
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "修图功能正在开发中，敬请期待！")
	default:
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "更多功能正在开发中，敬请期待！")
	}
}
