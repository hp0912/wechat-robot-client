package plugins

import (
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
	"wechat-robot-client/utils"
)

type AIChatPlugin struct{}

func NewAIChatPlugin() plugin.MessageHandler {
	return &AIChatPlugin{}
}

func (p *AIChatPlugin) GetName() string {
	return "AIChat"
}

func (p *AIChatPlugin) GetLabels() []string {
	return []string{"text", "internal", "chat"}
}

func (p *AIChatPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AIChatPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AIChatPlugin) Run(ctx *plugin.MessageContext) bool {
	aiTriggerWord := ctx.Settings.GetAITriggerWord()
	aiContext, err := ctx.MessageService.GetAIMessageContext(ctx.Message)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return true
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
	// 去除触发词后，剩下的内容如果为空，则不进行AI聊天
	if len(aiContext) > 0 {
		lastMessage := aiContext[len(aiContext)-1]
		if lastMessage.Content == "" && len(lastMessage.MultiContent) == 0 {
			if ctx.Message.IsChatRoom {
				ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "在呢", ctx.Message.SenderWxID)
			} else {
				ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "在呢")
			}
			return true
		}
	}
	aiChatService := service.NewAIChatService(ctx.Context, ctx.Settings)
	aiReply, err := aiChatService.Chat(aiContext)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return true
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
	return true
}
