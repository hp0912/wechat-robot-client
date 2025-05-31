package common_cron

import (
	"context"
	"log"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type ChatRoomRankingWeeklyCron struct {
	CronManager *CronManager
}

func NewChatRoomRankingWeeklyCron(cronManager *CronManager) vars.CommonCronInstance {
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

func (cron *ChatRoomRankingWeeklyCron) Cron() error {
	return service.NewChatRoomService(context.Background()).ChatRoomRankingWeekly()
}

func (cron *ChatRoomRankingWeeklyCron) Register() {
	if !cron.IsActive() {
		log.Println("每周群聊排行榜任务未启用")
		return
	}
	err := cron.CronManager.AddJob(vars.ChatRoomRankingWeeklyCron, *cron.CronManager.globalSettings.ChatRoomRankingWeeklyCron, func() {
		log.Println("开始执行每周群聊排行榜任务")
		if err := cron.Cron(); err != nil {
			log.Printf("每周群聊排行榜任务执行失败: %v", err)
		} else {
			log.Println("每周群聊排行榜任务执行完成")
		}
	})
	if err != nil {
		log.Printf("每周群聊排行榜任务注册失败: %v", err)
		return
	}
	log.Println("每周群聊排行榜任务初始化成功")
}
