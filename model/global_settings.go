package model

import (
	"database/sql"

	"gorm.io/datatypes"
)

type GlobalSettings struct {
	ID                        uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:公共配置表主键ID"`
	Owner                     string         `gorm:"column:owner;type:varchar(64);default:'';comment:所有者微信ID"`
	ChatAIEnabled             bool           `gorm:"column:chat_ai_enabled;default:false;comment:是否启用AI聊天功能"`
	ChatAITrigger             string         `gorm:"column:chat_ai_trigger;type:varchar(20);default:'';comment:触发聊天AI的关键词"`
	ChatBaseURL               string         `gorm:"column:chat_base_url;type:varchar(255);default:'';comment:聊天AI的基础URL地址"`
	ChatAPIKey                string         `gorm:"column:chat_api_key;type:varchar(255);default:'';comment:聊天AI的API密钥"`
	ChatModel                 string         `gorm:"column:chat_model;type:varchar(100);default:'';comment:聊天AI使用的模型名称"`
	ChatPrompt                sql.NullString `gorm:"column:chat_prompt;type:text;comment:聊天AI系统提示词"`
	ImageAIEnabled            bool           `gorm:"column:image_ai_enabled;default:false;comment:是否启用AI绘图功能"`
	ImageModel                string         `gorm:"column:image_model;type:varchar(255);default:'';comment:绘图AI模型"`
	ImageAISettings           datatypes.JSON `gorm:"column:image_ai_settings;type:json;comment:绘图AI配置项"`
	WelcomeEnabled            bool           `gorm:"column:welcome_enabled;default:false;comment:是否启用新成员加群欢迎功能"`
	WelcomeType               string         `gorm:"column:welcome_type;type:enum('text','emoji','image','url');default:'text';comment:欢迎方式：text-文本，emoji-表情，image-图片，url-链接"`
	WelcomeText               string         `gorm:"column:welcome_text;type:varchar(255);default:'';comment:欢迎新成员的文本"`
	WelcomeEmojiMD5           string         `gorm:"column:welcome_emoji_md5;type:varchar(64);default:'';comment:欢迎新成员的表情MD5"`
	WelcomeEmojiLen           int64          `gorm:"column:welcome_emoji_len;default:0;comment:欢迎新成员的表情MD5长度"`
	WelcomeImageURL           string         `gorm:"column:welcome_image_url;type:varchar(255);default:'';comment:欢迎新成员的图片URL"`
	WelcomeURL                string         `gorm:"column:welcome_url;type:varchar(255);default:'';comment:欢迎新成员的URL"`
	ChatRoomRankingEnabled    bool           `gorm:"column:chat_room_ranking_enabled;default:false;comment:是否启用群聊排行榜功能"`
	ChatRoomRankingDailyCron  string         `gorm:"column:chat_room_ranking_daily_cron;type:varchar(255);default:'';comment:每日定时任务表达式"`
	ChatRoomRankingWeeklyCron string         `gorm:"column:chat_room_ranking_weekly_cron;type:varchar(255);default:'';comment:每周定时任务表达式"`
	ChatRoomRankingMonthCron  string         `gorm:"column:chat_room_ranking_month_cron;type:varchar(255);default:'';comment:每月定时任务表达式"`
	ChatRoomSummaryEnabled    bool           `gorm:"column:chat_room_summary_enabled;default:false;comment:是否启用聊天记录总结功能"`
	ChatRoomSummaryModel      string         `gorm:"column:chat_room_summary_model;type:varchar(100);default:'';comment:聊天总结使用的AI模型名称"`
	ChatRoomSummaryCron       string         `gorm:"column:chat_room_summary_cron;type:varchar(100);default:'';comment:群聊总结的定时任务表达式"`
	NewsEnabled               bool           `gorm:"column:news_enabled;default:false;comment:是否启用每日早报功能"`
	NewsType                  string         `gorm:"column:news_type;type:enum('text','image');default:'text';comment:是否启用每日早报功能"`
	NewsCron                  string         `gorm:"column:news_cron;type:varchar(100);default:'';comment:每日早报的定时任务表达式"`
	MorningEnabled            bool           `gorm:"column:morning_enabled;default:false;comment:是否启用早安问候功能"`
	MorningCron               string         `gorm:"column:morning_cron;type:varchar(100);default:'';comment:早安问候的定时任务表达式"`
	FriendSyncCron            string         `gorm:"column:friend_sync_cron;type:varchar(100);default:'';comment:好友同步的定时任务表达式"`
}

// TableName 设置表名
func (GlobalSettings) TableName() string {
	return "global_settings"
}
