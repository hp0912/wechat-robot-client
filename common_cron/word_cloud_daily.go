package common_cron

import (
	"context"
	"log"
	"time"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type WordCloudDailyCron struct {
	CronManager *CronManager
}

func NewWordCloudDailyCron(cronManager *CronManager) vars.CommonCronInstance {
	return &WordCloudDailyCron{
		CronManager: cronManager,
	}
}

func (cron *WordCloudDailyCron) IsActive() bool {
	if cron.CronManager.globalSettings.ChatRoomRankingEnabled != nil && *cron.CronManager.globalSettings.ChatRoomRankingEnabled {
		return true
	}
	return false
}

func (cron *WordCloudDailyCron) Register() {
	if !cron.IsActive() {
		log.Println("每日词云任务未启用")
		return
	}
	if vars.WordCloudUrl == "" {
		log.Println("词云api地址未配置，无法执行每日词云任务")
		return
	}
	// 写死 5 0 * * *
	cron.CronManager.AddJob(vars.WordCloudDailyCron, "5 0 * * *", func(params ...any) error {
		log.Println("开始执行每日词云任务")

		// 获取今天凌晨零点
		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		// 获取昨天凌晨零点
		yesterdayStart := todayStart.AddDate(0, 0, -1)
		// 转换为时间戳（秒）
		yesterdayStartTimestamp := yesterdayStart.Unix()
		todayStartTimestamp := todayStart.Unix()

		wcService := service.NewWordCloudService(context.Background())
		settings := service.NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
		for _, setting := range settings {
			if setting == nil || setting.ChatRoomRankingEnabled == nil || !*setting.ChatRoomRankingEnabled {
				log.Printf("[词云] 群聊 %s 未开启群聊排行榜，跳过处理\n", setting.ChatRoomID)
				continue
			}
			imageData, err := wcService.WordCloudDaily(setting.ChatRoomID, yesterdayStartTimestamp, todayStartTimestamp)
			if err != nil {
				log.Printf("[词云] 群聊 %s 生成词云失败: %v\n", setting.ChatRoomID, err)
				continue
			}
			if imageData == nil {
				log.Printf("[词云] 群聊 %s 生成了空图片，跳过处理\n", setting.ChatRoomID)
				continue
			}

		}

		return nil
	})
	log.Println("每日词云任务初始化成功")
}
