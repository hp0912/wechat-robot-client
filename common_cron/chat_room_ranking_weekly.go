package common_cron

import (
	"log"
	"wechat-robot-client/vars"
)

type ChatRoomRankingWeeklyCron struct {
	CronManager *CronManager
}

func NewChatRoomRankingWeeklyCron(cronManager *CronManager) *ChatRoomRankingWeeklyCron {
	return &ChatRoomRankingWeeklyCron{
		CronManager: cronManager,
	}
}

func (cron *ChatRoomRankingWeeklyCron) IsActive() bool {
	if cron.CronManager.globalSettings.ChatRoomRankingEnabled != nil && *cron.CronManager.globalSettings.ChatRoomRankingEnabled {
		if cron.CronManager.globalSettings.ChatRoomRankingWeeklyCron != nil && *cron.CronManager.globalSettings.ChatRoomRankingWeeklyCron != "" {
			return true
		}
	}
	return false
}

func (cron *ChatRoomRankingWeeklyCron) Register() {
	if !cron.IsActive() {
		log.Println("每周群聊排行榜任务未启用")
		return
	}
	cron.CronManager.AddJob(vars.ChatRoomRankingWeeklyCron, *cron.CronManager.globalSettings.ChatRoomRankingWeeklyCron, func(params ...any) error {
		log.Println("开始执行每周群聊排行榜任务")
		return nil
	})
	log.Println("每周群聊排行榜任务初始化成功")
}
