package plugins

import (
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
	"wechat-robot-client/utils"
)

func OnChatIntention(ctx *plugin.MessageContext) {
	aiWorkflowService := service.NewAIWorkflowService(ctx.Context, ctx.Settings)
	aiTriggerWord := ctx.Settings.GetAITriggerWord()
	messageContent := ctx.MessageContent
	if ctx.Message.IsChatRoom {
		// 去除群聊中的AI触发词
		messageContent = utils.TrimAITriggerAll(messageContent, aiTriggerWord)
	}

	chatIntention := aiWorkflowService.ChatIntention(messageContent, ctx.ReferMessage)

	switch chatIntention {
	case service.ChatIntentionChat:
		OnAIChat(ctx)
	case service.ChatIntentionSing:
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "唱歌功能正在开发中，敬请期待！")
	case service.ChatIntentionSongRequest:
		title := aiWorkflowService.GetSongRequestTitle(messageContent)
		if title == "" {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "抱歉，我无法识别您想要点的歌曲。")
			return
		}
		err := ctx.MessageService.SendMusicMessage(ctx.Message.FromWxID, title)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		}
	case service.ChatIntentionDrawAPicture:
		isAIEnabled := ctx.Settings.IsAIDrawingEnabled()
		if !isAIEnabled {
			return
		}
		ctx.MessageContent = aiWorkflowService.GetDrawingPrompt(messageContent)
		OnAIDrawing(ctx)
	case service.ChatIntentionImageRecognizer:
		// 如果AI闲聊已经开启，则AI图片识别默认开启，AI聊天模型要支持多模态
		OnAIImageRecognizer(ctx)
	case service.ChatIntentionTTS:
		isTTSEnabled := ctx.Settings.IsTTSEnabled()
		if !isTTSEnabled {
			return
		}
		OnTTS(ctx)
	case service.ChatIntentionLTTS:
		isTTSEnabled := ctx.Settings.IsTTSEnabled()
		if !isTTSEnabled {
			return
		}
		OnLTTS(ctx)
	case service.ChatIntentionEditPictures:
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "修图功能正在开发中，敬请期待！")
	default:
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "更多功能正在开发中，敬请期待！")
	}
}
