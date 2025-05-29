package common_cron

import (
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"
)

type ChatRoomSummaryCron struct {
	CronManager    *CronManager
	GlobalSettings *model.GlobalSettings
}

func NewChatRoomSummaryCron(cronManager *CronManager, globalSettings *model.GlobalSettings) *ChatRoomSummaryCron {
	return &ChatRoomSummaryCron{
		CronManager:    cronManager,
		GlobalSettings: globalSettings,
	}
}

func (cron *ChatRoomSummaryCron) IsActive() bool {
	if cron.GlobalSettings.ChatRoomSummaryEnabled != nil && *cron.GlobalSettings.ChatRoomSummaryEnabled {
		return true
	}
	return false
}

func (cron *ChatRoomSummaryCron) Register() {
	if !cron.IsActive() {
		return
	}
	cron.CronManager.AddJob(vars.ChatRoomSummaryCron, cron.GlobalSettings.ChatRoomSummaryCron, func(params ...any) error {
		log.Println("开始执行每日群聊总结任务")
		return nil
	})
	log.Println("每日群聊总结任务初始化成功")
}
