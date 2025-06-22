package plugins

import (
	"encoding/base64"
	"fmt"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"

	"github.com/sashabaranov/go-openai"
)

func OnAIImageRecognizer(ctx *plugin.MessageContext) {
	// 下载引用的图片
	attachDownloadService := service.NewAttachDownloadService(ctx.Context)
	imageBytes, contentType, _, err := attachDownloadService.DownloadImage(ctx.ReferMessage.ID)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return
	}
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, base64Image)
	aiContext := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: dataURL,
					},
				},
				{
					Type: openai.ChatMessagePartTypeText,
					Text: ctx.MessageContent,
				},
			},
		},
	}
	aiChatService := service.NewAIChatService(ctx.Context, ctx.Settings)
	aiReply, err := aiChatService.Chat(aiContext)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return
	}
	// 当前场景只会返回文本，因此只提取文本
	var aiReplyText string
	if aiReply.Content != "" {
		aiReplyText = aiReply.Content
	} else if len(aiReply.MultiContent) > 0 {
		aiReplyText = aiReply.MultiContent[0].Text
	}
	if aiReplyText == "" {
		aiReplyText = "AI识别结果为空，请检查图片或重试。"
	}
	if ctx.Message.IsChatRoom {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReplyText, ctx.Message.SenderWxID)
	} else {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReplyText)
	}
}
