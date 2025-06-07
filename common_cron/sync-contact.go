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

func (cron *SyncContactCron) Cron() error {
	return service.NewContactService(context.Background()).SyncContact(true)
}

func (cron *SyncContactCron) Register() {
	if !cron.IsActive() {
		log.Println("联系人同步任务未启用")
		return
	}
	err := cron.CronManager.AddJob(vars.FriendSyncCron, cron.CronManager.globalSettings.FriendSyncCron, func() {
		log.Println("开始同步联系人")
		if err := cron.Cron(); err != nil {
			log.Printf("同步联系人失败: %v", err)
		} else {
			log.Println("联系人同步完成")
		}
	})
	if err != nil {
		log.Printf("联系人同步任务注册失败: %v", err)
		return
	}
	log.Println("同步联系人任务初始化成功")
}
