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
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return
		}
		err = pkg.SendDrawingImage(ctx.MessageService, ctx.Message.FromWxID, imageUrl)
		if err != nil {
			log.Printf("发送豆包图像失败: %v", err)
			return
		}
	case model.ImageModelGLM:
		// Handle 智谱模型
		var glmConfig pkg.GLMConfig
		if err := json.Unmarshal(aiConfig.ImageAISettings, &glmConfig); err != nil {
			log.Printf("反序列化智谱绘图配置失败: %v", err)
			return
		}
		glmConfig.Prompt = ctx.Message.Content
		imageUrl, err := pkg.GLM(&glmConfig)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return
		}
		err = pkg.SendDrawingImage(ctx.MessageService, ctx.Message.FromWxID, imageUrl)
		if err != nil {
			log.Printf("发送智谱图像失败: %v", err)
			return
		}
	case model.ImageModelHunyuan:
		// Handle 混元模型
		var hunyuanConfig pkg.HunyuanConfig
		if err := json.Unmarshal(aiConfig.ImageAISettings, &hunyuanConfig); err != nil {
			log.Printf("反序列化混元绘图配置失败: %v", err)
			return
		}
		hunyuanConfig.Prompt = ctx.Message.Content
		imageUrl, err := pkg.SubmitHunyuan(&hunyuanConfig)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return
		}
		err = pkg.SendDrawingImage(ctx.MessageService, ctx.Message.FromWxID, imageUrl)
		if err != nil {
			log.Printf("发送混元图像失败: %v", err)
			return
		}
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
