package global_cron

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/go-co-op/gocron"
)

type GlobalCron string

const (
	ChatRoomRankingDailyCron  GlobalCron = "chat_room_ranking_daily_cron"
	ChatRoomRankingWeeklyCron GlobalCron = "chat_room_ranking_weekly_cron"
	ChatRoomRankingMonthCron  GlobalCron = "chat_room_ranking_month_cron"
	ChatRoomSummaryCron       GlobalCron = "chat_room_summary_cron"
	NewsCron                  GlobalCron = "news_cron"
	MorningCron               GlobalCron = "morning_cron"
	FriendSyncCron            GlobalCron = "friend_sync_cron"
)

type TaskHandler func(params ...any) error

type CronManager struct {
	scheduler *gocron.Scheduler
	jobs      map[GlobalCron]*gocron.Job
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewCronManager() *CronManager {
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
	respo := repository.NewGlobalSettingsRepo(context.Background(), vars.DB)
	// 为空的时候，是从未扫码登陆的时候
	if vars.RobotRuntime.WxID != "" {
		globalSettings := respo.GetByOwner(vars.RobotRuntime.WxID)
		// 为 nil 的时候，是从未扫码登陆的时候
		if globalSettings != nil {

		}
	}

	// // 更新好友列表
	// _, _ = s.Cron("0 */1 * * *").Do(service.NewContactService(context.Background()).SyncContact(true))
	// // 开启定时任务
	// s.StartAsync()
	// log.Println("定时任务初始化成功")
}

func (m *CronManager) AddJob(cronName GlobalCron, cronExpr string, handler TaskHandler, params ...any) error {
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

func (m *CronManager) RemoveJob(cronName GlobalCron) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.jobs[cronName]; exists {
		m.scheduler.RemoveByTag(fmt.Sprintf("job_%s", cronName))
		delete(m.jobs, cronName)
		log.Printf("Job %s removed", cronName)
	}
	return nil
}

func (m *CronManager) UpdateJob(cronName GlobalCron, cronExpr string, handler TaskHandler, params ...any) error {
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
