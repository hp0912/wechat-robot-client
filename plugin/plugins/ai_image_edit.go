package plugins

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/plugin/pkg"
	"wechat-robot-client/service"

	doubaoModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type AImageEditPlugin struct{}

func NewAImageEditPlugin() plugin.MessageHandler {
	return &AImageEditPlugin{}
}

func (p *AImageEditPlugin) GetName() string {
	return "AImageEdit"
}

func (p *AImageEditPlugin) GetLabels() []string {
	return []string{"text", "internal", "image_edit"}
}

func (p *AImageEditPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AImageEditPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AImageEditPlugin) GetOSSFileURL(ctx *plugin.MessageContext) (string, error) {
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

func (p *AImageEditPlugin) Run(ctx *plugin.MessageContext) bool {
	aiConfig := ctx.Settings.GetAIConfig()
	switch aiConfig.ImageModel {
	case model.ImageModelDoubao:
		// Handle 豆包模型
		var doubaoConfig pkg.DoubaoConfig
		if err := json.Unmarshal(aiConfig.ImageAISettings, &doubaoConfig); err != nil {
			log.Printf("反序列化豆包绘图配置失败: %v", err)
			return true
		}
		if ctx.ReferMessage == nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "你需要引用一条图片消息。")
			return true
		}

		var dataURL string
		if ctx.ReferMessage.AttachmentUrl != "" {
			dataURL = ctx.ReferMessage.AttachmentUrl
		} else {
			imageURL, err := p.GetOSSFileURL(ctx)
			if err != nil {
				log.Printf("获取图片OSS URL失败: %v", err)
			}
			if imageURL != "" {
				dataURL = imageURL
			} else {
				attachDownloadService := service.NewAttachDownloadService(ctx.Context)
				imageBytes, contentType, _, err := attachDownloadService.DownloadImage(ctx.ReferMessage.ID)
				if err != nil {
					ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
					return true
				}
				base64Image := base64.StdEncoding.EncodeToString(imageBytes)
				dataURL = fmt.Sprintf("data:%s;base64,%s", contentType, base64Image)
			}
		}

		doubaoConfig.Model = "doubao-seededit-3-0-i2i-250628"
		doubaoConfig.Image = dataURL
		doubaoConfig.Prompt = ctx.MessageContent
		doubaoConfig.Size = doubaoModel.GenerateImagesSizeAdaptive
		doubaoConfig.Watermark = false
		imageUrl, err := pkg.DoubaoDrawing(&doubaoConfig)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return true
		}
		err = pkg.SendImageByURL(ctx.MessageService, ctx.Message.FromWxID, imageUrl)
		if err != nil {
			log.Printf("发送豆包图像失败: %v", err)
			return true
		}
	case model.ImageModelJimeng:
		// Handle 即梦模型
	case model.ImageModelGLM:
		// Handle 智谱模型
	case model.ImageModelHunyuan:
		// Handle 混元模型
	case model.ImageModelStableDiffusion:
		// Handle Stable Diffusion 模型
	case model.ImageModelMidjourney:
		// Handle Midjourney 模型
	case model.ImageModelOpenAI:
		// Handle OpenAI 模型
	default:
		log.Println("不支持的 AI 图像模型:", aiConfig.ImageModel)
	}
	return true
}
