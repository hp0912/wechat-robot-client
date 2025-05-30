package common_cron

import (
	"context"
	"log"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type SyncContactCron struct {
	CronManager *CronManager
}

func NewSyncContactCron(cronManager *CronManager) vars.CommonCronInstance {
	return &SyncContactCron{
		CronManager: cronManager,
	}
}

func (cron *SyncContactCron) IsActive() bool {
	return true
}

func (cron *SyncContactCron) Register() {
	if !cron.IsActive() {
		log.Println("联系人同步任务未启用")
		return
	}
	cron.CronManager.AddJob(vars.FriendSyncCron, cron.CronManager.globalSettings.FriendSyncCron, func(params ...any) error {
		log.Println("开始同步联系人")
		return service.NewContactService(context.Background()).SyncContact(true)
	})
	log.Println("同步联系人任务初始化成功")
}
