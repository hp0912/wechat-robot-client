package model

import (
	"gorm.io/datatypes"
)

type ChatRoomSettings struct {
	ID                        uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:公共配置表主键ID" json:"id"`
	ChatRoomID                string         `gorm:"column:chat_room_id;type:varchar(64);default:'';comment:群聊ID" json:"chat_room_id"`
	ChatAIEnabled             *bool          `gorm:"column:chat_ai_enabled;default:false;comment:是否启用AI聊天功能" json:"chat_ai_enabled"`
	ChatAITrigger             *string        `gorm:"column:chat_ai_trigger;type:varchar(20);default:'';comment:触发聊天AI的关键词" json:"chat_ai_trigger"`
	ChatBaseURL               *string        `gorm:"column:chat_base_url;type:varchar(255);default:'';comment:聊天AI的基础URL地址" json:"chat_base_url"`
	ChatAPIKey                *string        `gorm:"column:chat_api_key;type:varchar(255);default:'';comment:聊天AI的API密钥" json:"chat_api_key"`
	ChatModel                 *string        `gorm:"column:chat_model;type:varchar(100);default:'';comment:聊天AI使用的模型名称" json:"chat_model"`
	ImageRecognitionModel     *string        `gorm:"column:image_recognition_model;type:varchar(100);default:'';comment:图像识别AI使用的模型名称" json:"image_recognition_model"`
	ChatPrompt                *string        `gorm:"column:chat_prompt;type:text;comment:聊天AI系统提示词" json:"chat_prompt"`
	MaxCompletionTokens       *int           `gorm:"column:max_completion_tokens;default:0;comment:最大回复" json:"max_completion_tokens"`
	ImageAIEnabled            *bool          `gorm:"column:image_ai_enabled;default:false;comment:是否启用AI绘图功能" json:"image_ai_enabled"`
	ImageAISettings           datatypes.JSON `gorm:"column:image_ai_settings;type:json;comment:绘图AI配置项" json:"image_ai_settings"`
	TTSEnabled                *bool          `gorm:"column:tts_enabled;default:false;comment:是否启用AI文本转语音功能" json:"tts_enabled"`
	TTSSettings               datatypes.JSON `gorm:"column:tts_settings;type:json;comment:文本转语音配置项" json:"tts_settings"`
	LTTSSettings              datatypes.JSON `gorm:"column:ltts_settings;type:json;comment:长文本转语音配置项" json:"ltts_settings"`
	ShortVideoParsingEnabled  *bool          `gorm:"column:short_video_parsing_enabled;default:true;comment:是否启用短视频解析功能" json:"short_video_parsing_enabled"`
	PatEnabled                *bool          `gorm:"column:pat_enabled;default:false;comment:是否启用拍一拍功能" json:"pat_enabled"`
	PatType                   PatType        `gorm:"column:pat_type;type:enum('text','voice');default:'text';comment:拍一拍方式：text-文本，voice-语音" json:"pat_type"`
	PatText                   string         `gorm:"column:pat_text;type:varchar(255);default:'';comment:拍一拍的文本" json:"pat_text"`
	PatVoiceTimbre            string         `gorm:"column:pat_voice_timbre;type:varchar(255);default:'';comment:拍一拍的音色" json:"pat_voice_timbre"`
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
	ChatRoomSummaryEnabled    *bool          `gorm:"column:chat_room_summary_enabled;default:false;comment:是否启用聊天记录总结功能" json:"chat_room_summary_enabled"`
	ChatRoomSummaryModel      *string        `gorm:"column:chat_room_summary_model;type:varchar(100);default:'';comment:聊天总结使用的AI模型名称" json:"chat_room_summary_model"`
	NewsEnabled               *bool          `gorm:"column:news_enabled;default:false;comment:是否启用每日早报功能" json:"news_enabled"`
	NewsType                  *NewsType      `gorm:"column:news_type;type:enum('text','image');default:'text';comment:是否启用每日早报功能" json:"news_type"`
	MorningEnabled            *bool          `gorm:"column:morning_enabled;default:false;comment:是否启用早安问候功能" json:"morning_enabled"`
}

// TableName 设置表名
func (ChatRoomSettings) TableName() string {
	return "chat_room_settings"
}
