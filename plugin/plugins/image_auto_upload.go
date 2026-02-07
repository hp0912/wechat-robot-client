package plugins

import (
	"log"
	"time"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type ImageAutoUploadPlugin struct{}

func NewImageAutoUploadPlugin() plugin.MessageHandler {
	return &ImageAutoUploadPlugin{}
}

func (p *ImageAutoUploadPlugin) GetName() string {
	return "ImageAutoUpload"
}

func (p *ImageAutoUploadPlugin) GetLabels() []string {
	return []string{"image", "oss"}
}

func (p *ImageAutoUploadPlugin) Match(ctx *plugin.MessageContext) bool {
	return ctx.Message != nil && ctx.Message.Type == model.MsgTypeImage
}

func (p *ImageAutoUploadPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *ImageAutoUploadPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *ImageAutoUploadPlugin) Run(ctx *plugin.MessageContext) {
	if time.Now().Unix()-vars.RobotRuntime.LoginTime < 60 {
		log.Printf("登录时间不足60秒，跳过图片自动上传")
		return
	}
	ossSettingService := service.NewOSSSettingService(ctx.Context)
	ossSettings, err := ossSettingService.GetOSSSettingService()
	if err != nil {
		log.Printf("获取OSS设置失败: %v", err)
		return
	}
	if ossSettings == nil {
		log.Printf("OSS设置为空")
		return
	}
	if ossSettings.AutoUploadImage != nil && *ossSettings.AutoUploadImage && ossSettings.AutoUploadImageMode == model.AutoUploadModeAll {
		err := ossSettingService.UploadImageToOSS(ossSettings, ctx.Message)
		if err != nil {
			log.Printf("上传图片到OSS失败: %v", err)
		}
		return
	}
}
