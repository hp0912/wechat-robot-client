package common_cron

import (
	"context"
	"log"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

type GroupChatKnowledgeCron struct {
	CronManager *CronManager
}

func NewGroupChatKnowledgeCron(cronManager *CronManager) vars.CommonCronInstance {
	return &GroupChatKnowledgeCron{
		CronManager: cronManager,
	}
}

func (cron *GroupChatKnowledgeCron) IsActive() bool {
	return vars.KnowledgeService != nil
}

func (cron *GroupChatKnowledgeCron) Cron() error {
	return service.NewGroupChatKnowledgeService(context.Background()).ExtractGroupChatKnowledge()
}

func (cron *GroupChatKnowledgeCron) Register() {
	if !cron.IsActive() {
		return
	}
	err := cron.CronManager.AddJob(vars.GroupChatKnowledgeCron, "0 * * * *", func() {
		if err := cron.Cron(); err != nil {
			log.Printf("[GroupChatKnowledgeCron] 执行失败: %v", err)
		}
	})
	if err != nil {
		log.Printf("[GroupChatKnowledgeCron] 注册定时任务失败: %v", err)
	}
}
