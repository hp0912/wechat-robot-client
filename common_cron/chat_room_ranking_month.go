package common_cron

import (
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"
)

type ChatRoomRankingMonthCron struct {
	CronManager    *CronManager
	GlobalSettings *model.GlobalSettings
}

func NewChatRoomRankingMonthCron(cronManager *CronManager, globalSettings *model.GlobalSettings) *ChatRoomRankingMonthCron {
	return &ChatRoomRankingMonthCron{
		CronManager:    cronManager,
		GlobalSettings: globalSettings,
	}
}

func (cron *ChatRoomRankingMonthCron) IsActive() bool {
	if cron.GlobalSettings.ChatRoomRankingEnabled != nil && *cron.GlobalSettings.ChatRoomRankingEnabled {
		if cron.GlobalSettings.ChatRoomRankingMonthCron != nil && *cron.GlobalSettings.ChatRoomRankingMonthCron != "" {
			return true
		}
	}
	return false
}

func (cron *ChatRoomRankingMonthCron) Start() {
	if !cron.IsActive() {
		return
	}
	cron.CronManager.AddJob(vars.ChatRoomRankingMonthCron, *cron.GlobalSettings.ChatRoomRankingMonthCron, func(params ...any) error {
		log.Println("开始执行每月群聊排行榜任务")
		return nil
	})
	log.Println("每月群聊排行榜任务初始化成功")
}
