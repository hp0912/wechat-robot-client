package global_cron

import (
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"
)

type ChatRoomRankingDailyCron struct {
	CronManager    *CronManager
	GlobalSettings *model.GlobalSettings
}

func NewChatRoomRankingDailyCron(cronManager *CronManager, globalSettings *model.GlobalSettings) *ChatRoomRankingDailyCron {
	return &ChatRoomRankingDailyCron{
		CronManager:    cronManager,
		GlobalSettings: globalSettings,
	}
}

func (cron *ChatRoomRankingDailyCron) IsActive() bool {
	if cron.GlobalSettings.ChatRoomRankingEnabled != nil && *cron.GlobalSettings.ChatRoomRankingEnabled {
		return true
	}
	return false
}

func (cron *ChatRoomRankingDailyCron) Start() {
	if !cron.IsActive() {
		return
	}
	cron.CronManager.AddJob(vars.ChatRoomRankingDailyCron, cron.GlobalSettings.ChatRoomRankingDailyCron, func(params ...any) error {
		log.Println("开始执行每日群聊排行榜任务")
		return nil
	})
	log.Println("每日群聊排行榜任务初始化成功")
}
