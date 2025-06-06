package common_cron

import (
	"context"
	"log"
	"wechat-robot-client/service"
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

func (cron *ChatRoomRankingMonthCron) Cron() error {
	return service.NewChatRoomService(context.Background()).ChatRoomRankingMonthly()
}

func (cron *ChatRoomRankingMonthCron) Register() {
	if !cron.IsActive() {
		log.Println("每月群聊排行榜任务未启用")
		return
	}
	err := cron.CronManager.AddJob(vars.ChatRoomRankingMonthCron, *cron.CronManager.globalSettings.ChatRoomRankingMonthCron, func() {
		log.Println("开始执行每月群聊排行榜任务")
		if err := cron.Cron(); err != nil {
			log.Printf("每月群聊排行榜任务执行失败: %v", err)
		} else {
			log.Println("每月群聊排行榜任务执行完成")
		}
	})
	if err != nil {
		log.Printf("每月群聊排行榜任务注册失败: %v", err)
		return
	}
	log.Println("每月群聊排行榜任务初始化成功")
}
