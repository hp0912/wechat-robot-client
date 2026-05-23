package plugins

import (
	"fmt"
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/service"
)

type AIVoiceUploadPlugin struct{}

func NewAIVoiceUploadPlugin() plugin.MessageHandler {
	return &AIVoiceUploadPlugin{}
}

func (p *AIVoiceUploadPlugin) GetName() string {
	return "AIVoiceUpload"
}

func (p *AIVoiceUploadPlugin) GetLabels() []string {
	return []string{"text", "internal", "chat"}
}

func (p *AIVoiceUploadPlugin) Match(ctx *plugin.MessageContext) bool {
	return ctx.ReferMessage != nil
}

func (p *AIVoiceUploadPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AIVoiceUploadPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AIVoiceUploadPlugin) GetOSSFileURL(ctx *plugin.MessageContext) (string, error) {
	ossSettingService := service.NewOSSSettingService(ctx.Context)
	ossSettings, err := ossSettingService.GetOSSSettingService()
	if err != nil {
		return "", fmt.Errorf("获取OSS设置失败: %w", err)
	}
	if ossSettings.AutoUploadVoice != nil && *ossSettings.AutoUploadVoice {
		err := ossSettingService.UploadVoiceToOSS(ossSettings, ctx.ReferMessage)
		if err != nil {
			return "", fmt.Errorf("上传语音到OSS失败: %w", err)
		}
		return ctx.ReferMessage.AttachmentUrl, nil
	}
	return "", nil
}

func (p *AIVoiceUploadPlugin) SendMessage(ctx *plugin.MessageContext, aiReplyText string) {
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

func (p *AIVoiceUploadPlugin) Run(ctx *plugin.MessageContext) {
	if ctx.ReferMessage.AttachmentUrl == "" {
		voiceURL, err := p.GetOSSFileURL(ctx)
		if err != nil {
			log.Printf("语音上传失败: %v", err)
			p.SendMessage(ctx, fmt.Sprintf("语音上传失败: %v，你可能没开启自动上传语音，请前往机器人详情 -> OSS 设置手动开启", err))
			return
		}
		if voiceURL == "" {
			p.SendMessage(ctx, "语音上传失败: 语音URL为空，你可能没开启自动上传语音，请前往机器人详情 -> OSS 设置手动开启")
			return
		}
	}
}
