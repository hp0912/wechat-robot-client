package global_cron

import (
	"context"
	"log"
	"wechat-robot-client/model"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type SyncContactCron struct {
	CronManager    *CronManager
	GlobalSettings *model.GlobalSettings
}

func NewSyncContactCron(cronManager *CronManager, globalSettings *model.GlobalSettings) *SyncContactCron {
	return &SyncContactCron{
		CronManager:    cronManager,
		GlobalSettings: globalSettings,
	}
}

func (cron *SyncContactCron) IsActive() bool {
	return true
}

func (cron *SyncContactCron) Start() {
	if !cron.IsActive() {
		log.Println("联系人同步任务未启用")
		return
	}
	cron.CronManager.AddJob(vars.FriendSyncCron, cron.GlobalSettings.FriendSyncCron, func(params ...any) error {
		log.Println("开始同步联系人")
		return service.NewContactService(context.Background()).SyncContact(true)
	})
	log.Println("同步联系人任务初始化成功")
}
