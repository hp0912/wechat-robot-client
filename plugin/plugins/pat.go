package plugins

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/plugin/pkg"
)

type PatPlugin struct{}

func NewPatPlugin() plugin.MessageHandler {
	return &PatPlugin{}
}

func (p *PatPlugin) GetName() string {
	return "Pat"
}

func (p *PatPlugin) GetLabels() []string {
	return []string{"pat"}
}

func (p *PatPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *PatPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *PatPlugin) Run(ctx *plugin.MessageContext) bool {
	if !ctx.Pat {
		return false
	}
	patConfig := ctx.Settings.GetPatConfig()
	if !patConfig.PatEnabled {
		return true
	}
	if patConfig.PatType == model.PatTypeText {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, patConfig.PatText)
		return true
	}
	if patConfig.PatType == model.PatTypeVoice {
		isTTSEnabled := ctx.Settings.IsTTSEnabled()
		if !isTTSEnabled {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "文本转语音功能未开启，请联系管理员。")
			return true
		}
		aiConfig := ctx.Settings.GetAIConfig()
		var doubaoConfig pkg.DoubaoTTSConfig
		if err := json.Unmarshal(aiConfig.TTSSettings, &doubaoConfig); err != nil {
			log.Printf("反序列化豆包文本转语音配置失败: %v", err)
			return true
		}
		doubaoConfig.RequestBody.ReqParams.Speaker = patConfig.PatVoiceTimbre
		doubaoConfig.RequestBody.ReqParams.Text = patConfig.PatText

		audioBase64, err := pkg.DoubaoTTSSubmit(&doubaoConfig)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("豆包文本转语音请求失败: %v", err), ctx.Message.SenderWxID)
			return true
		}
		audioData, err := base64.StdEncoding.DecodeString(audioBase64)
		if err != nil {
			ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("音频数据解码失败: %v", err), ctx.Message.SenderWxID)
			return true
		}
		audioReader := bytes.NewReader(audioData)
		ctx.MessageService.MsgSendVoice(ctx.Message.FromWxID, audioReader, fmt.Sprintf(".%s", doubaoConfig.RequestBody.ReqParams.AudioParams.Format))

		return true
	}
	return true
}
