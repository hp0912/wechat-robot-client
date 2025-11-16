package plugins

import (
	"fmt"
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type AIImageUploadPlugin struct{}

func NewAIImageUploadPlugin() plugin.MessageHandler {
	return &AIImageUploadPlugin{}
}

func (p *AIImageUploadPlugin) GetName() string {
	return "AIImageUpload"
}

func (p *AIImageUploadPlugin) GetLabels() []string {
	return []string{"text", "internal", "chat"}
}

func (p *AIImageUploadPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AIImageUploadPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AIImageUploadPlugin) GetOSSFileURL(ctx *plugin.MessageContext) (string, error) {
	ossSettingService := service.NewOSSSettingService(ctx.Context)
	ossSettings, err := ossSettingService.GetOSSSettingService()
	if err != nil {
		return "", fmt.Errorf("获取OSS设置失败: %w", err)
	}
	if ossSettings == nil {
		return "", fmt.Errorf("OSS设置为空")
	}
	if ossSettings.AutoUploadImage != nil && *ossSettings.AutoUploadImage {
		ossSettingService := service.NewOSSSettingService(ctx.Context)
		err := ossSettingService.UploadImageToOSS(ossSettings, ctx.ReferMessage)
		if err != nil {
			return "", fmt.Errorf("上传图片到OSS失败: %w", err)
		}
		return ctx.ReferMessage.AttachmentUrl, nil
	}
	return "", nil
}

func (p *AIImageUploadPlugin) SendMessage(ctx *plugin.MessageContext, aiReplyText string) {
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

func (p *AIImageUploadPlugin) Run(ctx *plugin.MessageContext) bool {
	if ctx.ReferMessage == nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "你需要引用一条图片消息。")
		return true
	}
	if ctx.ReferMessage.AttachmentUrl == "" {
		imageURL, err := p.GetOSSFileURL(ctx)
		if err != nil {
			log.Printf("图片上传失败: %v", err)
			p.SendMessage(ctx, fmt.Sprintf("图片上传失败: %v，你可能没开启自动上传图片，请前往机器人详情 -> OSS 设置手动开启", err))
			return true
		}
		if imageURL == "" {
			p.SendMessage(ctx, "图片上传失败: 图片URL为空，你可能没开启自动上传图片，请前往机器人详情 -> OSS 设置手动开启")
			return true
		}
	}

	p.SendMessage(ctx, "已经收到您的图片啦～")

	return true
}
