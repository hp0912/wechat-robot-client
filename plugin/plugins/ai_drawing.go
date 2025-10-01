package plugins

import (
	"encoding/json"
	"log"
	"strings"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/plugin/pkg"
)

type AIDrawingPlugin struct{}

func NewAIDrawingPlugin() plugin.MessageHandler {
	return &AIDrawingPlugin{}
}

func (p *AIDrawingPlugin) GetName() string {
	return "AIDrawing"
}

func (p *AIDrawingPlugin) GetLabels() []string {
	return []string{"text", "internal", "drawing"}
}

func (p *AIDrawingPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AIDrawingPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AIDrawingPlugin) Run(ctx *plugin.MessageContext) bool {
	aiConfig := ctx.Settings.GetAIConfig()
	switch aiConfig.ImageModel {
	case model.ImageModelDoubao:
		// Handle 豆包模型
		var doubaoConfig pkg.DoubaoConfig
		if err := json.Unmarshal(aiConfig.ImageAISettings, &doubaoConfig); err != nil {
			log.Printf("反序列化豆包绘图配置失败: %v", err)
			return true
		}
		doubaoConfig.Prompt = ctx.MessageContent
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
		var jimengConfig pkg.JimengConfig
		if err := json.Unmarshal(aiConfig.ImageAISettings, &jimengConfig); err != nil {
			log.Printf("反序列化即梦绘图配置失败: %v", err)
			return true
		}
		jimengConfig.Prompt = ctx.MessageContent
		imageUrl, err := pkg.JimengDrawing(&jimengConfig)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return true
		}
		imageUrls := strings.Split(imageUrl, "\n")
		for _, imgurl := range imageUrls {
			if imgurl == "" {
				continue
			}
			err = pkg.SendImageByURL(ctx.MessageService, ctx.Message.FromWxID, imgurl)
			if err != nil {
				log.Printf("发送即梦图像失败: %v", err)
				return true
			}
		}
	case model.ImageModelGLM:
		// Handle 智谱模型
		var glmConfig pkg.GLMConfig
		if err := json.Unmarshal(aiConfig.ImageAISettings, &glmConfig); err != nil {
			log.Printf("反序列化智谱绘图配置失败: %v", err)
			return true
		}
		glmConfig.Prompt = ctx.MessageContent
		imageUrl, err := pkg.GLMDrawing(&glmConfig)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return true
		}
		err = pkg.SendImageByURL(ctx.MessageService, ctx.Message.FromWxID, imageUrl)
		if err != nil {
			log.Printf("发送智谱图像失败: %v", err)
			return true
		}
	case model.ImageModelHunyuan:
		// Handle 混元模型
		var hunyuanConfig pkg.HunyuanConfig
		if err := json.Unmarshal(aiConfig.ImageAISettings, &hunyuanConfig); err != nil {
			log.Printf("反序列化混元绘图配置失败: %v", err)
			return true
		}
		hunyuanConfig.Prompt = ctx.MessageContent
		imageUrl, err := pkg.SubmitHunyuanDrawing(&hunyuanConfig)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, err.Error())
			return true
		}
		err = pkg.SendImageByURL(ctx.MessageService, ctx.Message.FromWxID, imageUrl)
		if err != nil {
			log.Printf("发送混元图像失败: %v", err)
			return true
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
	return true
}
