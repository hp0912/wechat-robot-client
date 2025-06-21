package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"wechat-robot-client/interface/plugin"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/distributedlock"
	"wechat-robot-client/plugin/pkg"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
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

	// 解析引用消息
	doubaoConfig.Text = ""

	lockCtx := context.Background()
	lock := distributedlock.NewDistributedLock(vars.RedisClient, fmt.Sprintf("doubao_ltts_lock:%s", ctx.Message.FromWxID),
		distributedlock.WithExpiration(10*time.Second),
		distributedlock.WithMaxRetries(5),
		distributedlock.WithAutoRenewal(),
	)
	if err := lock.Lock(lockCtx); err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "操作太快了，休息一下吧", ctx.Message.SenderWxID)
		return
	}
	defer lock.Unlock(lockCtx)

	aiTaskService := service.NewAITaskService(ctx.Context)
	// 查询是否存在进行中的任务
	existingTask, err := aiTaskService.GetOngoingByWeChatID(ctx.Message.FromWxID)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("查询进行中的长文本转语音任务失败: %v", err), ctx.Message.SenderWxID)
		return
	}
	if len(existingTask) > 0 {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, "您有进行中的长文本转语音任务，请等待任务完成后再提交新的任务", ctx.Message.SenderWxID)
		return
	}

	taskID, err := pkg.DoubaoLTTSSubmit(&doubaoConfig)
	if err != nil {
		ctx.MessageService.SendTextMessage(ctx.Message.FromWxID, fmt.Sprintf("豆包长文本转语音任务提交失败: %v", err), ctx.Message.SenderWxID)
		return
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
		return
	}

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
}
