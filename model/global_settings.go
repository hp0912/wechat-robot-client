package model

import (
	"gorm.io/datatypes"
)

type WelcomeType string

const (
	WelcomeTypeText  WelcomeType = "text"  // 文本
	WelcomeTypeEmoji WelcomeType = "emoji" // 表情
	WelcomeTypeImage WelcomeType = "image" // 图片
	WelcomeTypeURL   WelcomeType = "url"   // 链接
)

type NewsType string

const (
	NewsTypeNone  NewsType = ""
	NewsTypeText  NewsType = "text"  // 文本
	NewsTypeImage NewsType = "image" // 图片
)

type ImageModel string

const (
	ImageModelDoubao          ImageModel = "doubao"           // 豆包模型
	ImageModelJimeng          ImageModel = "jimeng"           // 即梦模型
	ImageModelGLM             ImageModel = "glm"              // 智谱模型
	ImageModelHunyuan         ImageModel = "hunyuan"          // 腾讯混元模型
	ImageModelStableDiffusion ImageModel = "stable-diffusion" // Stable Diffusion 模型
	ImageModelMidjourney      ImageModel = "midjourney"       // Midjourney 模型
	ImageModelOpenAI          ImageModel = "openai"           // OpenAI 模型
)

type GlobalSettings struct {
	ID                        uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:公共配置表主键ID" json:"id"`
	ChatAIEnabled             *bool          `gorm:"column:chat_ai_enabled;default:false;comment:是否启用AI聊天功能" json:"chat_ai_enabled"`
	ChatAITrigger             *string        `gorm:"column:chat_ai_trigger;type:varchar(20);default:'';comment:触发聊天AI的关键词" json:"chat_ai_trigger"`
	ChatBaseURL               string         `gorm:"column:chat_base_url;type:varchar(255);default:'';comment:聊天AI的基础URL地址" json:"chat_base_url"`
	ChatAPIKey                string         `gorm:"column:chat_api_key;type:varchar(255);default:'';comment:聊天AI的API密钥" json:"chat_api_key"`
	ChatModel                 string         `gorm:"column:chat_model;type:varchar(100);default:'';comment:聊天AI使用的模型名称" json:"chat_model"`
	ChatPrompt                string         `gorm:"column:chat_prompt;type:text;comment:聊天AI系统提示词" json:"chat_prompt"`
	ImageAIEnabled            *bool          `gorm:"column:image_ai_enabled;default:false;comment:是否启用AI绘图功能" json:"image_ai_enabled"`
	ImageModel                ImageModel     `gorm:"column:image_model;type:varchar(255);default:'';comment:绘图AI模型" json:"image_model"`
	ImageAISettings           datatypes.JSON `gorm:"column:image_ai_settings;type:json;comment:绘图AI配置项" json:"image_ai_settings"`
	TTSEnabled                *bool          `gorm:"column:tts_enabled;default:false;comment:是否启用AI文本转语音功能" json:"tts_enabled"`
	TTSSettings               datatypes.JSON `gorm:"column:tts_settings;type:json;comment:文本转语音配置项" json:"tts_settings"`
	LTTSSettings              datatypes.JSON `gorm:"column:ltts_settings;type:json;comment:长文本转语音配置项" json:"ltts_settings"`
	WelcomeEnabled            *bool          `gorm:"column:welcome_enabled;default:false;comment:是否启用新成员加群欢迎功能" json:"welcome_enabled"`
	WelcomeType               WelcomeType    `gorm:"column:welcome_type;type:enum('text','emoji','image','url');default:'text';comment:欢迎方式：text-文本，emoji-表情，image-图片，url-链接" json:"welcome_type"`
	WelcomeText               string         `gorm:"column:welcome_text;type:varchar(255);default:'';comment:欢迎新成员的文本" json:"welcome_text"`
	WelcomeEmojiMD5           string         `gorm:"column:welcome_emoji_md5;type:varchar(64);default:'';comment:欢迎新成员的表情MD5" json:"welcome_emoji_md5"`
	WelcomeEmojiLen           int64          `gorm:"column:welcome_emoji_len;default:0;comment:欢迎新成员的表情MD5长度" json:"welcome_emoji_len"`
	WelcomeImageURL           string         `gorm:"column:welcome_image_url;type:varchar(255);default:'';comment:欢迎新成员的图片URL" json:"welcome_image_url"`
	WelcomeURL                string         `gorm:"column:welcome_url;type:varchar(255);default:'';comment:欢迎新成员的URL" json:"welcome_url"`
	LeaveChatRoomAlertEnabled *bool          `gorm:"column:leave_chat_room_alert_enabled;default:false;comment:是否启用离开群聊提醒功能" json:"leave_chat_room_alert_enabled"`
	LeaveChatRoomAlertText    string         `gorm:"column:leave_chat_room_alert_text;type:varchar(255);default:'';comment:离开群聊提醒文本" json:"leave_chat_room_alert_text"`
	ChatRoomRankingEnabled    *bool          `gorm:"column:chat_room_ranking_enabled;default:false;comment:是否启用群聊排行榜功能" json:"chat_room_ranking_enabled"`
	ChatRoomRankingDailyCron  string         `gorm:"column:chat_room_ranking_daily_cron;type:varchar(255);default:'';comment:每日定时任务表达式" json:"chat_room_ranking_daily_cron"`
	ChatRoomRankingWeeklyCron *string        `gorm:"column:chat_room_ranking_weekly_cron;type:varchar(255);default:'';comment:每周定时任务表达式" json:"chat_room_ranking_weekly_cron"`
	ChatRoomRankingMonthCron  *string        `gorm:"column:chat_room_ranking_month_cron;type:varchar(255);default:'';comment:每月定时任务表达式" json:"chat_room_ranking_month_cron"`
	ChatRoomSummaryEnabled    *bool          `gorm:"column:chat_room_summary_enabled;default:false;comment:是否启用聊天记录总结功能" json:"chat_room_summary_enabled"`
	ChatRoomSummaryModel      string         `gorm:"column:chat_room_summary_model;type:varchar(100);default:'';comment:聊天总结使用的AI模型名称" json:"chat_room_summary_model"`
	ChatRoomSummaryCron       string         `gorm:"column:chat_room_summary_cron;type:varchar(100);default:'';comment:群聊总结的定时任务表达式" json:"chat_room_summary_cron"`
	NewsEnabled               *bool          `gorm:"column:news_enabled;default:false;comment:是否启用每日早报功能" json:"news_enabled"`
	NewsType                  NewsType       `gorm:"column:news_type;type:enum('text','image');default:'text';comment:是否启用每日早报功能" json:"news_type"`
	NewsCron                  string         `gorm:"column:news_cron;type:varchar(100);default:'';comment:每日早报的定时任务表达式" json:"news_cron"`
	MorningEnabled            *bool          `gorm:"column:morning_enabled;default:false;comment:是否启用早安问候功能" json:"morning_enabled"`
	MorningCron               string         `gorm:"column:morning_cron;type:varchar(100);default:'';comment:早安问候的定时任务表达式" json:"morning_cron"`
	FriendSyncCron            string         `gorm:"column:friend_sync_cron;type:varchar(100);default:'';comment:好友同步的定时任务表达式" json:"friend_sync_cron"`
}

// TableName 设置表名
func (GlobalSettings) TableName() string {
	return "global_settings"
}
