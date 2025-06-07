package vars

import (
	"context"
	"wechat-robot-client/model"
)

type CommonCron string

const (
	WordCloudDailyCron        CommonCron = "word_cloud_daily_cron"
	ChatRoomRankingDailyCron  CommonCron = "chat_room_ranking_daily_cron"
	ChatRoomRankingWeeklyCron CommonCron = "chat_room_ranking_weekly_cron"
	ChatRoomRankingMonthCron  CommonCron = "chat_room_ranking_month_cron"
	ChatRoomSummaryCron       CommonCron = "chat_room_summary_cron"
	NewsCron                  CommonCron = "news_cron"
	MorningCron               CommonCron = "morning_cron"
	FriendSyncCron            CommonCron = "friend_sync_cron"
)

type TaskHandler func()

type CronManagerInterface interface {
	Name() string
	Shutdown(ctx context.Context) error
	SetGlobalSettings(globalSettings *model.GlobalSettings)
	Start()
	AddJob(cronName CommonCron, cronExpr string, handler TaskHandler) error
	RemoveJob(cronName CommonCron) error
	UpdateJob(cronName CommonCron, cronExpr string, handler TaskHandler) error
	Clear()
	Stop()
}

type CommonCronInstance interface {
	IsActive() bool
	Cron() error
	Register()
}
