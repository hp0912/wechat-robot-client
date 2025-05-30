package common_cron

import (
	"log"
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
	cron.CronManager.AddJob(vars.ChatRoomSummaryCron, cron.CronManager.globalSettings.ChatRoomSummaryCron, func(params ...any) error {
		log.Println("开始执行每日群聊总结任务")
		return nil
	})
	log.Println("每日群聊总结任务初始化成功")
}
