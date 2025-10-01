package plugins

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
	"unicode/utf8"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/distributedlock"
	"wechat-robot-client/plugin/pkg"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type AITTSPlugin struct{}

func NewAITTSPlugin() plugin.MessageHandler {
	return &AITTSPlugin{}
}

func (p *AITTSPlugin) GetName() string {
	return "AITTS"
}

func (p *AITTSPlugin) GetLabels() []string {
	return []string{"text", "internal", "voice"}
}

func (p *AITTSPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AITTSPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AITTSPlugin) Run(ctx *plugin.MessageContext) bool {
	var ttsContent string
	if ctx.Message.Type == model.MsgTypeText {
		aiWorkflowService := service.NewAIWorkflowService(ctx.Context, ctx.Settings)
		ttsContent = aiWorkflowService.GetTTSText(ctx.MessageContent)
	}
	if ctx.Message.Type == model.MsgTypeApp && ctx.Message.AppMsgType == model.AppMsgTypequote && ctx.ReferMessage.Type == model.MsgTypeText {
		ttsContent = ctx.ReferMessage.Content
	}
	if ttsContent == "" {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "未提前到有效的文本内容", ctx.Message.SenderWxID)
		return true
	}
	if utf8.RuneCountInString(ttsContent) > 260 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "你要说的也太多了，要不你还是说点别的吧。", ctx.Message.SenderWxID)
		return true
	}
	aiConfig := ctx.Settings.GetAIConfig()
	var doubaoConfig pkg.DoubaoTTSConfig
	if err := json.Unmarshal(aiConfig.TTSSettings, &doubaoConfig); err != nil {
		log.Printf("反序列化豆包文本转语音配置失败: %v", err)
		return true
	}
	doubaoConfig.Request.Text = ttsContent

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
	ctx.MessageService.MsgSendVoice(ctx.Message.FromWxID, audioReader, fmt.Sprintf(".%s", doubaoConfig.Audio.Encoding))

	return true
}

type AILTTSPlugin struct{}

func NewAILTTSPlugin() plugin.MessageHandler {
	return &AILTTSPlugin{}
}

func (p *AILTTSPlugin) GetName() string {
	return "AILTTS"
}

func (p *AILTTSPlugin) GetLabels() []string {
	return []string{"internal", "voice"}
}

func (p *AILTTSPlugin) PreAction(ctx *plugin.MessageContext) bool {
	return true
}

func (p *AILTTSPlugin) PostAction(ctx *plugin.MessageContext) {

}

func (p *AILTTSPlugin) Run(ctx *plugin.MessageContext) bool {
	if ctx.ReferMessage == nil {
		log.Println("长文本转语音引用消息为空")
		return true
	}
	referXml, err := ctx.MessageService.XmlDecoder(ctx.ReferMessage.Content)
	if err != nil {
		log.Printf("解析引用消息失败: %v", err)
		return true
	}
	if referXml.AppMsg.Type != 6 || referXml.AppMsg.AppAttach.FileExt != "txt" {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "请使用TXT文本文件进行长文本转语音", ctx.Message.SenderWxID)
		return true
	}

	aiConfig := ctx.Settings.GetAIConfig()
	var doubaoConfig pkg.DoubaoLTTSConfig
	if err := json.Unmarshal(aiConfig.LTTSSettings, &doubaoConfig); err != nil {
		log.Printf("反序列化豆包长文本转语音配置失败: %v", err)
		return true
	}

	// 解析引用消息的文本文件内容
	reader, _, err := service.NewAttachDownloadService(ctx.Context).DownloadFile(ctx.ReferMessage.ID)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("下载引用消息的文本文件失败: %v", err), ctx.Message.SenderWxID)
		return true
	}
	defer reader.Close()
	// 读取文本内容
	textBytes, err := io.ReadAll(reader)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("读取文本文件内容失败: %v", err), ctx.Message.SenderWxID)
		return true
	}
	doubaoConfig.Text = string(textBytes)

	lockCtx := context.Background()
	lock := distributedlock.NewDistributedLock(vars.RedisClient, fmt.Sprintf("doubao_ltts_lock:%s", ctx.Message.SenderWxID),
		distributedlock.WithExpiration(10*time.Second),
		distributedlock.WithMaxRetries(5),
		distributedlock.WithAutoRenewal(),
	)
	if err := lock.Lock(lockCtx); err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "操作太快了，休息一下吧", ctx.Message.SenderWxID)
		return true
	}
	defer lock.Unlock(lockCtx)

	aiTaskService := service.NewAITaskService(ctx.Context)
	// 查询是否存在进行中的任务
	existingTask, err := aiTaskService.GetOngoingByWeChatID(ctx.Message.FromWxID)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("查询进行中的长文本转语音任务失败: %v", err), ctx.Message.SenderWxID)
		return true
	}
	if len(existingTask) > 0 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "您有进行中的长文本转语音任务，请等待任务完成后再提交新的任务", ctx.Message.SenderWxID)
		return true
	}

	taskID, err := pkg.DoubaoLTTSSubmit(&doubaoConfig)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("豆包长文本转语音任务提交失败: %v", err), ctx.Message.SenderWxID)
		return true
	}
	aiTask := model.AITask{
		ContactID:        ctx.Message.FromWxID,
		MessageID:        ctx.Message.ID,
		AIProviderTaskID: taskID,
		AITaskType:       model.AITaskTypeLongTextTTS,
		AITaskStatus:     model.AITaskStatusProcessing,
		Extra:            nil, // 可以根据需要填充额外信息
		CreatedAt:        time.Now().Unix(),
		UpdatedAt:        time.Now().Unix(),
	}
	err = aiTaskService.CreateAITask(&aiTask)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("创建长文本转语音任务失败: %v", err), ctx.Message.SenderWxID)
		return true
	}
	ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "您的长文本转语音任务已提交，请耐心等待。", ctx.Message.SenderWxID)
	doubaoConfig.TaskID = taskID

	go func() {
		pollInterval := 5 * time.Minute // 每5分钟轮询一次
		maxDuration := 3 * time.Hour    // 最大等待3小时
		startTime := time.Now()
		for {
			audioURL, err := pkg.DoubaoLTTSQuery(&doubaoConfig)
			if err != nil {
				// 先查询一遍，因为任务可能已经被回调接口处理过了
				_aiTask, _ := aiTaskService.GetByID(aiTask.ID)
				if _aiTask != nil && (_aiTask.AITaskStatus == model.AITaskStatusPending || _aiTask.AITaskStatus == model.AITaskStatusProcessing) {
					ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("豆包长文本转语音任务失败: %v", err), ctx.Message.SenderWxID)
				}
				aiTask.AITaskStatus = model.AITaskStatusFailed
				aiTask.UpdatedAt = time.Now().Unix()
				_ = aiTaskService.UpdateAITask(&aiTask)
				return
			}
			if audioURL != "" {
				// 先查询一遍，因为任务可能已经被回调接口处理过了
				_aiTask, _ := aiTaskService.GetByID(aiTask.ID)
				if _aiTask != nil && (_aiTask.AITaskStatus == model.AITaskStatusPending || _aiTask.AITaskStatus == model.AITaskStatusProcessing) {
					ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("豆包长文本转语音任务完成，有效期24小时，请及时下载，音频地址: %s", audioURL), ctx.Message.SenderWxID)
				}
				aiTask.AITaskStatus = model.AITaskStatusCompleted
				extraData := map[string]string{
					"audio_url": audioURL,
				}
				extraJSON, _ := json.Marshal(extraData)
				aiTask.Extra = extraJSON
				aiTask.UpdatedAt = time.Now().Unix()
				_ = aiTaskService.UpdateAITask(&aiTask)
				return
			}
			if time.Since(startTime) >= maxDuration {
				ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "豆包长文本转语音任务超时，请稍后再试", ctx.Message.SenderWxID)
				return
			}
			time.Sleep(pollInterval)
		}
	}()

	return true
}
