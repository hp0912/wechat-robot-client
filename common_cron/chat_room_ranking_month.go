package common_cron

import (
	"log"
	"wechat-robot-client/vars"
)

type ChatRoomRankingMonthCron struct {
	CronManager *CronManager
}

func NewChatRoomRankingMonthCron(cronManager *CronManager) vars.CommonCronInstance {
	return &ChatRoomRankingMonthCron{
		CronManager: cronManager,
	}
}

func (cron *ChatRoomRankingMonthCron) IsActive() bool {
	if cron.CronManager.globalSettings.ChatRoomRankingEnabled != nil && *cron.CronManager.globalSettings.ChatRoomRankingEnabled {
		if cron.CronManager.globalSettings.ChatRoomRankingMonthCron != nil && *cron.CronManager.globalSettings.ChatRoomRankingMonthCron != "" {
			return true
		}
	}
	return false
}

func (cron *ChatRoomRankingMonthCron) Register() {
	if !cron.IsActive() {
		log.Println("每月群聊排行榜任务未启用")
		return
	}
	cron.CronManager.AddJob(vars.ChatRoomRankingMonthCron, *cron.CronManager.globalSettings.ChatRoomRankingMonthCron, func(params ...any) error {
		log.Println("开始执行每月群聊排行榜任务")
		return nil
	})
	log.Println("每月群聊排行榜任务初始化成功")
}
