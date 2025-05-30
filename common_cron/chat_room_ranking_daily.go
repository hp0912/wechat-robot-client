package common_cron

import (
	"log"
	"wechat-robot-client/vars"
)

type ChatRoomRankingDailyCron struct {
	CronManager *CronManager
}

func NewChatRoomRankingDailyCron(cronManager *CronManager) vars.CommonCronInstance {
	return &ChatRoomRankingDailyCron{
		CronManager: cronManager,
	}
}

func (cron *ChatRoomRankingDailyCron) IsActive() bool {
	if cron.CronManager.globalSettings.ChatRoomRankingEnabled != nil && *cron.CronManager.globalSettings.ChatRoomRankingEnabled {
		return true
	}
	return false
}

func (cron *ChatRoomRankingDailyCron) Register() {
	if !cron.IsActive() {
		log.Println("每日群聊排行榜任务未启用")
		return
	}
	cron.CronManager.AddJob(vars.ChatRoomRankingDailyCron, cron.CronManager.globalSettings.ChatRoomRankingDailyCron, func(params ...any) error {
		log.Println("开始执行每日群聊排行榜任务")
		return nil
	})
	log.Println("每日群聊排行榜任务初始化成功")
}
