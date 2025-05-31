package common_cron

import (
	"context"
	"log"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type ChatRoomSummaryCron struct {
	CronManager *CronManager
}

func NewChatRoomSummaryCron(cronManager *CronManager) vars.CommonCronInstance {
	return &ChatRoomSummaryCron{
		CronManager: cronManager,
	}
}

func (cron *ChatRoomSummaryCron) IsActive() bool {
	if cron.CronManager.globalSettings.ChatRoomSummaryEnabled != nil && *cron.CronManager.globalSettings.ChatRoomSummaryEnabled {
		return true
	}
	return false
}

func (cron *ChatRoomSummaryCron) Register() {
	if !cron.IsActive() {
		log.Println("每日群聊总结任务未启用")
		return
	}
	err := cron.CronManager.AddJob(vars.ChatRoomSummaryCron, cron.CronManager.globalSettings.ChatRoomSummaryCron, func() error {
		log.Println("开始执行每日群聊总结任务")
		return service.NewChatRoomService(context.Background()).ChatRoomAISummary()
	})
	if err != nil {
		log.Printf("每日群聊总结任务注册失败: %v", err)
		return
	}
	log.Println("每日群聊总结任务初始化成功")
}
