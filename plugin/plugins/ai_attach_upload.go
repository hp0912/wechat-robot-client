package plugins

import (
	"fmt"
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type AIAttachUploadPlugin struct{}

func NewAIAttachUploadPlugin() plugin.MessageHandler {
	return &AIAttachUploadPlugin{}
}

func (p *AIAttachUploadPlugin) GetName() string {
	return "AIAttachUpload"
}

func (p *AIAttachUploadPlugin) GetLabels() []string {
	return []string{"text", "internal", "chat"}
}

func (p *AIAttachUploadPlugin) Match(ctx *plugin.MessageContext) bool {
	return ctx.ReferMessage != nil
}

func (p *AIAttachUploadPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AIAttachUploadPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AIAttachUploadPlugin) GetOSSFileURL(ctx *plugin.MessageContext) (string, error) {
	ossSettingService := service.NewOSSSettingService(ctx.Context)
	ossSettings, err := ossSettingService.GetOSSSettingService()
	if err != nil {
		return "", fmt.Errorf("获取OSS设置失败: %w", err)
	}
	if ossSettings.AutoUploadFile != nil && *ossSettings.AutoUploadFile {
		err := ossSettingService.UploadFileToOSS(ossSettings, ctx.ReferMessage)
		if err != nil {
			return "", fmt.Errorf("上传文件到OSS失败: %w", err)
		}
		return ctx.ReferMessage.AttachmentUrl, nil
	}
	return "", nil
}

func (p *AIAttachUploadPlugin) SendMessage(ctx *plugin.MessageContext, aiReplyText string) {
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

func (p *AIAttachUploadPlugin) Run(ctx *plugin.MessageContext) {
	if ctx.ReferMessage.AttachmentUrl == "" {
		fileURL, err := p.GetOSSFileURL(ctx)
		if err != nil {
			log.Printf("文件上传失败: %v", err)
			p.SendMessage(ctx, fmt.Sprintf("文件上传失败: %v，你可能没开启自动上传文件，请前往机器人详情 -> OSS 设置手动开启", err))
			return
		}
		if fileURL == "" {
			p.SendMessage(ctx, "文件上传失败: 文件URL为空，你可能没开启自动上传文件，请前往机器人详情 -> OSS 设置手动开启")
			return
		}
	}
}
