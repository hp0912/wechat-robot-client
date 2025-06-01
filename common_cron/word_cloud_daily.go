package common_cron

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func (cron *WordCloudDailyCron) Cron() error {
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()
	// 创建词云缓存目录
	wordCloudCacheDir := filepath.Join(string(filepath.Separator), "app", "word_cloud_cache")
	if err := os.MkdirAll(wordCloudCacheDir, 0755); err != nil {
		log.Printf("创建词云缓存目录失败: %v", err)
		return err
	}
	wcService := service.NewWordCloudService(context.Background())
	settings, err := service.NewChatRoomSettingsService(context.Background()).GetAllEnableChatRank()
	if err != nil {
		log.Printf("获取群聊设置失败: %v", err)
		return err
	}
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
		// 保存词云图片
		dateStr := yesterdayStart.Format("2006-01-02")
		filename := fmt.Sprintf("%s_%s.png", setting.ChatRoomID, dateStr)
		filePath := filepath.Join(wordCloudCacheDir, filename)
		if err := os.WriteFile(filePath, imageData, 0644); err != nil {
			log.Printf("[词云] 群聊 %s 保存词云图片失败: %v\n", setting.ChatRoomID, err)
			continue
		}
		log.Printf("[词云] 群聊 %s 词云图片已保存至: %s\n", setting.ChatRoomID, filePath)
	}
	return nil
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
	err := cron.CronManager.AddJob(vars.WordCloudDailyCron, "5 0 * * *", func() {
		log.Println("开始执行每日词云任务")
		if err := cron.Cron(); err != nil {
			log.Printf("每日词云任务执行失败: %v", err)
		} else {
			log.Println("每日词云任务执行完成")
		}
	})
	if err != nil {
		log.Printf("每日词云任务注册失败: %v", err)
		return
	}
	log.Println("每日词云任务初始化成功")
}
