package plugins

import (
	"fmt"
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type AIEmojiUploadPlugin struct{}

func NewAIEmojiUploadPlugin() plugin.MessageHandler {
	return &AIEmojiUploadPlugin{}
}

func (p *AIEmojiUploadPlugin) GetName() string {
	return "AIEmojiUpload"
}

func (p *AIEmojiUploadPlugin) GetLabels() []string {
	return []string{"text", "internal", "chat"}
}

func (p *AIEmojiUploadPlugin) Match(ctx *plugin.MessageContext) bool {
	return ctx.ReferMessage != nil
}

func (p *AIEmojiUploadPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AIEmojiUploadPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AIEmojiUploadPlugin) GetOSSFileURL(ctx *plugin.MessageContext) (string, error) {
	ossSettingService := service.NewOSSSettingService(ctx.Context)
	ossSettings, err := ossSettingService.GetOSSSettingService()
	if err != nil {
		return "", fmt.Errorf("获取OSS设置失败: %w", err)
	}
	if ossSettings.AutoUploadImage != nil && *ossSettings.AutoUploadImage {
		err := ossSettingService.UploadEmojiToOSS(ossSettings, ctx.ReferMessage)
		if err != nil {
			return "", fmt.Errorf("上传表情包到OSS失败: %w", err)
		}
		return ctx.ReferMessage.AttachmentUrl, nil
	}
	return "", nil
}

func (p *AIEmojiUploadPlugin) SendMessage(ctx *plugin.MessageContext, aiReplyText string) {
	var err error
	if ctx.Message.IsChatRoom {
		err = ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReplyText, ctx.Message.SenderWxID)
	} else {
		err = ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, aiReplyText)
	}
	if err != nil {
		log.Printf("发送AI回复消息失败: %v", err)
	}
}

func (p *AIEmojiUploadPlugin) Run(ctx *plugin.MessageContext) {
	if ctx.ReferMessage.AttachmentUrl == "" {
		emojiURL, err := p.GetOSSFileURL(ctx)
		if err != nil {
			log.Printf("表情包上传失败: %v", err)
			p.SendMessage(ctx, fmt.Sprintf("表情包上传失败: %v，你可能没开启自动上传图片，请前往机器人详情 -> OSS 设置手动开启", err))
			return
		}
		if emojiURL == "" {
			p.SendMessage(ctx, "表情包上传失败: 表情包URL为空，你可能没开启自动上传图片，请前往机器人详情 -> OSS 设置手动开启")
			return
		}
	}
}
