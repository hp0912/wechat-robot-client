package common_cron

import (
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"
)

type ChatRoomRankingWeeklyCron struct {
	CronManager    *CronManager
	GlobalSettings *model.GlobalSettings
}

func NewChatRoomRankingWeeklyCron(cronManager *CronManager, globalSettings *model.GlobalSettings) *ChatRoomRankingWeeklyCron {
	return &ChatRoomRankingWeeklyCron{
		CronManager:    cronManager,
		GlobalSettings: globalSettings,
	}
}

func (cron *ChatRoomRankingWeeklyCron) IsActive() bool {
	if cron.GlobalSettings.ChatRoomRankingEnabled != nil && *cron.GlobalSettings.ChatRoomRankingEnabled {
		if cron.GlobalSettings.ChatRoomRankingWeeklyCron != nil && *cron.GlobalSettings.ChatRoomRankingWeeklyCron != "" {
			return true
		}
	}
	return false
}

func (cron *ChatRoomRankingWeeklyCron) Start() {
	if !cron.IsActive() {
		return
	}
	cron.CronManager.AddJob(vars.ChatRoomRankingWeeklyCron, *cron.GlobalSettings.ChatRoomRankingWeeklyCron, func(params ...any) error {
		log.Println("开始执行每周群聊排行榜任务")
		return nil
	})
	log.Println("每周群聊排行榜任务初始化成功")
}
