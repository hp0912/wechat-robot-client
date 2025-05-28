package global_cron

import (
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"
)

type GoodMorningCron struct {
	CronManager    *CronManager
	GlobalSettings *model.GlobalSettings
}

func NewGoodMorningCron(cronManager *CronManager, globalSettings *model.GlobalSettings) *GoodMorningCron {
	return &GoodMorningCron{
		CronManager:    cronManager,
		GlobalSettings: globalSettings,
	}
}

func (cron *GoodMorningCron) IsActive() bool {
	if cron.GlobalSettings.MorningEnabled != nil && *cron.GlobalSettings.MorningEnabled {
		return true
	}
	return false
}

func (cron *GoodMorningCron) Start() {
	if !cron.IsActive() {
		log.Println("每日早安任务未启用")
		return
	}
	cron.CronManager.AddJob(vars.FriendSyncCron, cron.GlobalSettings.FriendSyncCron, func(params ...any) error {
		log.Println("开始每日早安任务")
		return nil
	})
	log.Println("每日早安任务初始化成功")
}
