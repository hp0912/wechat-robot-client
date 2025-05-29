package vars

import "context"

type CommonCron string

const (
	ChatRoomRankingDailyCron  CommonCron = "chat_room_ranking_daily_cron"
	ChatRoomRankingWeeklyCron CommonCron = "chat_room_ranking_weekly_cron"
	ChatRoomRankingMonthCron  CommonCron = "chat_room_ranking_month_cron"
	ChatRoomSummaryCron       CommonCron = "chat_room_summary_cron"
	NewsCron                  CommonCron = "news_cron"
	MorningCron               CommonCron = "morning_cron"
	FriendSyncCron            CommonCron = "friend_sync_cron"
)

type TaskHandler func(params ...any) error

type CronManagerInterface interface {
	Name() string
	Shutdown(ctx context.Context) error
	Start()
	AddJob(cronName CommonCron, cronExpr string, handler TaskHandler, params ...any) error
	RemoveJob(cronName CommonCron) error
	UpdateJob(cronName CommonCron, cronExpr string, handler TaskHandler, params ...any) error
	Stop()
}
