package plugins

import (
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
	"wechat-robot-client/utils"
)

func OnAIChat(ctx *plugin.MessageContext) {
	aiTriggerWord := ctx.Settings.GetAITriggerWord()
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
	var aiReplyText string
	if aiReply.Content != "" {
		aiReplyText = aiReply.Content
	} else if len(aiReply.MultiContent) > 0 {
		aiReplyText = aiReply.MultiContent[0].Text
	}
	if aiReplyText == "" {
		aiReplyText = "AI返回了空内容。"
	}
	// 待处理，AI返回了图片

	if ctx.Message.IsChatRoom {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReplyText, ctx.Message.SenderWxID)
	} else {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReplyText)
	}
}
