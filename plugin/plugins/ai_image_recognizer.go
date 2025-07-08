package plugins

import (
	"encoding/base64"
	"fmt"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"

	"github.com/sashabaranov/go-openai"
)

type AImageRecognizerPlugin struct{}

func NewAImageRecognizerPlugin() plugin.MessageHandler {
	return &AImageRecognizerPlugin{}
}

func (p *AImageRecognizerPlugin) GetName() string {
	return "AImageRecognizer"
}

func (p *AImageRecognizerPlugin) GetLabels() []string {
	return []string{"internal", "chat"}
}

func (p *AImageRecognizerPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AImageRecognizerPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AImageRecognizerPlugin) Run(ctx *plugin.MessageContext) bool {
	if ctx.ReferMessage == nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "你需要引用一条图片消息。")
		return true
	}
	// 下载引用的图片
	attachDownloadService := service.NewAttachDownloadService(ctx.Context)
	imageBytes, contentType, _, err := attachDownloadService.DownloadImage(ctx.ReferMessage.ID)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
		return true
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
		return true
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
	return true
}
