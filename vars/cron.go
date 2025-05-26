package vars

import "context"

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

type CronManagerInterface interface {
	Name() string
	Shutdown(ctx context.Context) error
	Start()
	AddJob(cronName GlobalCron, cronExpr string, handler TaskHandler, params ...any) error
	RemoveJob(cronName GlobalCron) error
	UpdateJob(cronName GlobalCron, cronExpr string, handler TaskHandler, params ...any) error
	Stop()
}
