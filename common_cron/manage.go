package common_cron

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"wechat-robot-client/model"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"

	"github.com/go-co-op/gocron"
)

type CronManager struct {
	scheduler      *gocron.Scheduler
	isRunning      bool
	globalSettings *model.GlobalSettings
	jobs           map[vars.CommonCron]*gocron.Job
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

type CronInstance interface {
	IsActive() bool
	Register()
}

func NewCronManager() vars.CronManagerInterface {
	globalSettings, err := service.NewGlobalSettingsService(context.Background()).GetGlobalSettings()
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &CronManager{
		scheduler:      gocron.NewScheduler(time.Local),
		globalSettings: globalSettings,
		jobs:           make(map[vars.CommonCron]*gocron.Job),
		isRunning:      false,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (m *CronManager) Name() string {
	return "公共定时任务"
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

func (m *CronManager) SetGlobalSettings(globalSettings *model.GlobalSettings) {
	m.globalSettings = globalSettings
}

func (m *CronManager) Start() {
	// 启动调度器
	if !m.isRunning {
		m.scheduler.StartAsync()
		m.isRunning = true
	}
	// 为空的时候，是从未扫码登陆的时候
	if vars.RobotRuntime.WxID != "" {
		// 为 nil 的时候，是从未扫码登陆的时候
		if m.globalSettings != nil {
			// 同步联系人
			syncContactCron := NewSyncContactCron(m)
			syncContactCron.Register()
			// 每日早安
			morningCron := NewGoodMorningCron(m)
			morningCron.Register()
			// 每日早报
			newsCron := NewNewsCron(m)
			newsCron.Register()
			// 每日群聊总结
			chatRoomSummaryCron := NewChatRoomSummaryCron(m)
			chatRoomSummaryCron.Register()
			// 每日词云
			wordCloudDailyCron := NewWordCloudDailyCron(m)
			wordCloudDailyCron.Register()
			// 每日群聊排行榜
			chatRoomRankingDailyCron := NewChatRoomRankingDailyCron(m)
			chatRoomRankingDailyCron.Register()
			// 每周群聊排行榜
			chatRoomRankingWeeklyCron := NewChatRoomRankingWeeklyCron(m)
			chatRoomRankingWeeklyCron.Register()
			// 每月群聊排行榜
			chatRoomRankingMonthCron := NewChatRoomRankingMonthCron(m)
			chatRoomRankingMonthCron.Register()
		}
	}
}

func (m *CronManager) AddJob(cronName vars.CommonCron, cronExpr string, handler vars.TaskHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 添加到调度器
	cronJob, err := m.scheduler.Cron(cronExpr).Do(handler)
	if err != nil {
		return fmt.Errorf("failed to schedule job %s: %w", cronName, err)
	}
	// 设置任务标签
	cronJob.Tag(fmt.Sprintf("job_%s", cronName))
	m.jobs[cronName] = cronJob
	log.Printf("Job %s scheduled with cron: %s", cronName, cronExpr)
	return nil
}

func (m *CronManager) RemoveJob(cronName vars.CommonCron) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.jobs[cronName]; exists {
		m.scheduler.RemoveByTag(fmt.Sprintf("job_%s", cronName))
		delete(m.jobs, cronName)
		log.Printf("Job %s removed", cronName)
	}
	return nil
}

func (m *CronManager) UpdateJob(cronName vars.CommonCron, cronExpr string, handler vars.TaskHandler) error {
	// 先移除旧任务
	if err := m.RemoveJob(cronName); err != nil {
		return err
	}
	// 添加新任务
	return m.AddJob(cronName, cronExpr, handler)
}

func (m *CronManager) Clear() {
	m.scheduler.Clear()
}

func (m *CronManager) Stop() {
	m.cancel()
	if m.isRunning {
		m.scheduler.Stop()
		m.isRunning = false
	}
}
