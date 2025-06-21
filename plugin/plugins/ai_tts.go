package plugins

import (
	"encoding/json"
	"fmt"
	"log"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/plugin/pkg"
)

// OnTTS 文本转语音
func OnTTS(ctx *plugin.MessageContext) {
	// aiConfig := ctx.Settings.GetAIConfig()
}

// OnLTTS 长文本转语音
func OnLTTS(ctx *plugin.MessageContext) {
	aiConfig := ctx.Settings.GetAIConfig()
	var doubaoConfig pkg.DoubaoLTTSConfig
	if err := json.Unmarshal(aiConfig.LTTSSettings, &doubaoConfig); err != nil {
		log.Printf("反序列化豆包长文本转语音配置失败: %v", err)
		return
	}
	doubaoConfig.Text = ""
	taskID, err := pkg.DoubaoLTTSSubmit(&doubaoConfig)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("豆包长文本转语音任务提交失败: %v", err))
		return
	}
	doubaoConfig.TaskID = taskID
}
