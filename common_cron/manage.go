package common_cron

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"

	"github.com/go-co-op/gocron"
)

type CronManager struct {
	scheduler *gocron.Scheduler
	jobs      map[vars.GlobalCron]*gocron.Job
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

type CronInstance interface {
	IsActive() bool
	Start()
}

func NewCronManager() vars.CronManagerInterface {
	ctx, cancel := context.WithCancel(context.Background())
	return &CronManager{
		scheduler: gocron.NewScheduler(time.Local),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (m *CronManager) Name() string {
	return "全局定时任务"
}

func (m *CronManager) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		m.Stop()
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *CronManager) Start() {
	// 启动调度器
	m.scheduler.StartAsync()
	// 为空的时候，是从未扫码登陆的时候
	if vars.RobotRuntime.WxID != "" {
		globalSettings := service.NewGlobalSettingsService(context.Background()).GetGlobalSettings()
		// 为 nil 的时候，是从未扫码登陆的时候
		if globalSettings != nil {
			// 同步联系人
			syncContactCron := NewSyncContactCron(m, globalSettings)
			syncContactCron.Start()
			// 每日早安
			morningCron := NewGoodMorningCron(m, globalSettings)
			morningCron.Start()
			// 每日早报
			newsCron := NewNewsCron(m, globalSettings)
			newsCron.Start()
			// 每日群聊总结
			chatRoomSummaryCron := NewChatRoomSummaryCron(m, globalSettings)
			chatRoomSummaryCron.Start()
			// 每日群聊排行榜
			chatRoomRankingDailyCron := NewChatRoomRankingDailyCron(m, globalSettings)
			chatRoomRankingDailyCron.Start()
			// 每周群聊排行榜
			chatRoomRankingWeeklyCron := NewChatRoomRankingWeeklyCron(m, globalSettings)
			chatRoomRankingWeeklyCron.Start()
			// 每月群聊排行榜
			chatRoomRankingMonthCron := NewChatRoomRankingMonthCron(m, globalSettings)
			chatRoomRankingMonthCron.Start()
		}
	}
}

func (m *CronManager) AddJob(cronName vars.GlobalCron, cronExpr string, handler vars.TaskHandler, params ...any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 添加到调度器
	cronJob, err := m.scheduler.Cron(cronExpr).Do(handler, params...)
	if err != nil {
		return fmt.Errorf("failed to schedule job %s: %w", cronName, err)
	}
	// 设置任务标签
	cronJob.Tag(fmt.Sprintf("job_%s", cronName))
	m.jobs[cronName] = cronJob
	log.Printf("Job %s scheduled with cron: %s", cronName, cronExpr)
	return nil
}

func (m *CronManager) RemoveJob(cronName vars.GlobalCron) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.jobs[cronName]; exists {
		m.scheduler.RemoveByTag(fmt.Sprintf("job_%s", cronName))
		delete(m.jobs, cronName)
		log.Printf("Job %s removed", cronName)
	}
	return nil
}

func (m *CronManager) UpdateJob(cronName vars.GlobalCron, cronExpr string, handler vars.TaskHandler, params ...any) error {
	// 先移除旧任务
	if err := m.RemoveJob(cronName); err != nil {
		return err
	}
	// 添加新任务
	return m.AddJob(cronName, cronExpr, handler, params...)
}

func (m *CronManager) Stop() {
	m.cancel()
	m.scheduler.Stop()
}
