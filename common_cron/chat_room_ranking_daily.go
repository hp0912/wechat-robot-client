package common_cron

import (
	"context"
	"log"
	"wechat-robot-client/service"
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
	err := cron.CronManager.AddJob(vars.ChatRoomRankingDailyCron, cron.CronManager.globalSettings.ChatRoomRankingDailyCron, func() error {
		log.Println("开始执行每日群聊排行榜任务")
		return service.NewChatRoomService(context.Background()).ChatRoomRankingDaily()
	})
	if err != nil {
		log.Printf("每日群聊排行榜任务注册失败: %v", err)
		return
	}
	log.Println("每日群聊排行榜任务初始化成功")
}
