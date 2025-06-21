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

	chatIntention := aiWorkflowService.ChatIntention(messageContent)

	switch chatIntention {
	case service.ChatIntentionChat:
		aiContext, err := ctx.MessageService.GetAIMessageContext(ctx.Message)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return
		}
		if ctx.Message.IsChatRoom {
			for index := range aiContext {
				// 去除群聊中的AI触发词
				aiContext[index].Content = utils.TrimAITriggerAll(aiContext[index].Content, aiTriggerWord)
				for index2 := range aiContext[index].MultiContent {
					// 去除群聊中的AI触发词
					aiContext[index].MultiContent[index2].Text = utils.TrimAITriggerAll(aiContext[index].MultiContent[index2].Text, aiTriggerWord)
				}
			}
		}
		aiChatService := service.NewAIChatService(ctx.Context, ctx.Settings)
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
