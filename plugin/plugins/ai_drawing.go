package plugins

import (
	"encoding/json"
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/plugin/pkg"
)

func OnAIDrawing(ctx *plugin.MessageContext) {
	aiConfig := ctx.Settings.GetAIConfig()
	switch aiConfig.ImageModel {
	case model.ImageModelDoubao:
		// Handle 豆包模型
		var doubaoConfig pkg.DoubaoConfig
		if err := json.Unmarshal(aiConfig.ImageAISettings, &doubaoConfig); err != nil {
			log.Printf("反序列化豆包绘图配置失败: %v", err)
			return
		}
		doubaoConfig.Prompt = ctx.Message.Content
		imageUrl, err := pkg.Doubao(&doubaoConfig)
		if err != nil {
			log.Printf("生成豆包图像失败: %v", err)
			return
		}
		err = pkg.SendDrawingImage(ctx.MessageService, ctx.Message.FromWxID, imageUrl)
		if err != nil {
			log.Printf("发送豆包图像失败: %v", err)
			return
		}
	case model.ImageModelGLM:
		// Handle 智谱模型
	case model.ImageModelStableDiffusion:
		// Handle Stable Diffusion 模型
	case model.ImageModelMidjourney:
		// Handle Midjourney 模型
	case model.ImageModelOpenAI:
		// Handle OpenAI 模型
	default:
		log.Println("不支持的 AI 图像模型:", aiConfig.ImageModel)
	}
}
