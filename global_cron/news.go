package global_cron

import (
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"
)

type NewsCron struct {
	CronManager    *CronManager
	GlobalSettings *model.GlobalSettings
}

func NewNewsCron(cronManager *CronManager, globalSettings *model.GlobalSettings) *NewsCron {
	return &NewsCron{
		CronManager:    cronManager,
		GlobalSettings: globalSettings,
	}
}

func (cron *NewsCron) IsActive() bool {
	if cron.GlobalSettings.NewsEnabled != nil && *cron.GlobalSettings.NewsEnabled {
		return true
	}
	return false
}

func (cron *NewsCron) Start() {
	if !cron.IsActive() {
		return
	}
	cron.CronManager.AddJob(vars.FriendSyncCron, cron.GlobalSettings.FriendSyncCron, func(params ...any) error {
		log.Println("开始同步联系人")
		return nil
	})
	log.Println("同步联系人任务初始化成功")
}
