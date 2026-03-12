package common_cron

import (
	"log"
	"wechat-robot-client/vars"
)

type MemoryMaintenanceCron struct {
	CronManager *CronManager
}

func NewMemoryMaintenanceCron(cronManager *CronManager) vars.CommonCronInstance {
	return &MemoryMaintenanceCron{
		CronManager: cronManager,
	}
}

func (cron *MemoryMaintenanceCron) IsActive() bool {
	return vars.MemoryService != nil
}

func (cron *MemoryMaintenanceCron) Cron() error {
	// 衰减长期未访问记忆
	vars.MemoryService.DecayOldMemories()
	// 总结过期会话（10 分钟未活跃）
	vars.MemoryService.SummarizeExpiredSessions(10)
	return nil
}

func (cron *MemoryMaintenanceCron) Register() {
	if !cron.IsActive() {
		return
	}
	// 每 30 分钟执行一次会话总结 + 记忆衰减
	err := cron.CronManager.AddJob(vars.MemoryDecayCron, "*/30 * * * *", func() {
		if err := cron.Cron(); err != nil {
			log.Printf("[MemoryCron] 执行失败: %v", err)
		}
	})
	if err != nil {
		log.Printf("[MemoryCron] 注册定时任务失败: %v", err)
	}
}
