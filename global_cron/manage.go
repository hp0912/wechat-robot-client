package global_cron

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
			m.AddJob(vars.FriendSyncCron, globalSettings.FriendSyncCron, func(params ...any) error {
				log.Println("开始同步联系人")
				return service.NewContactService(context.Background()).SyncContact(true)
			})
			log.Println("同步联系人任务初始化成功")

			if globalSettings.MorningEnabled != nil && *globalSettings.MorningEnabled {
				m.AddJob(vars.MorningCron, globalSettings.MorningCron, func(params ...any) error {
					log.Println("开始执行每日早安任务")
					return nil
				})
				log.Println("每日早安任务初始化成功")
			}

			if globalSettings.NewsEnabled != nil && *globalSettings.NewsEnabled {
				m.AddJob(vars.NewsCron, globalSettings.NewsCron, func(params ...any) error {
					log.Println("开始执行每日早报任务")
					return nil
				})
				log.Println("每日早报任务初始化成功")
			}

			if globalSettings.ChatRoomSummaryEnabled != nil && *globalSettings.ChatRoomSummaryEnabled {
				m.AddJob(vars.ChatRoomSummaryCron, globalSettings.ChatRoomSummaryCron, func(params ...any) error {
					log.Println("开始执行每日群聊总结任务")
					return nil
				})
				log.Println("每日群聊总结任务初始化成功")
			}

			if globalSettings.ChatRoomRankingEnabled != nil && *globalSettings.ChatRoomRankingEnabled {
				m.AddJob(vars.ChatRoomRankingDailyCron, globalSettings.ChatRoomRankingDailyCron, func(params ...any) error {
					log.Println("开始执行每日群聊排行榜任务")
					return nil
				})
				log.Println("每日群聊排行榜任务初始化成功")

				if globalSettings.ChatRoomRankingWeeklyCron != nil && *globalSettings.ChatRoomRankingWeeklyCron != "" {
					m.AddJob(vars.ChatRoomRankingWeeklyCron, *globalSettings.ChatRoomRankingWeeklyCron, func(params ...any) error {
						log.Println("开始执行每周群聊排行榜任务")
						return nil
					})
					log.Println("每周群聊排行榜任务初始化成功")
				}

				if globalSettings.ChatRoomRankingMonthCron != nil && *globalSettings.ChatRoomRankingMonthCron != "" {
					m.AddJob(vars.ChatRoomRankingMonthCron, *globalSettings.ChatRoomRankingMonthCron, func(params ...any) error {
						log.Println("开始执行每月群聊排行榜任务")
						return nil
					})
					log.Println("每月群聊排行榜任务初始化成功")
				}
			}
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
