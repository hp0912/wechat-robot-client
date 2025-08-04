package model

import (
	"gorm.io/datatypes"
)

type FriendSettings struct {
	ID                    uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:公共配置表主键ID" json:"id"`
	WeChatID              string         `gorm:"column:wechat_id;type:varchar(64);default:'';comment:好友微信ID" json:"wechat_id"`
	ChatAIEnabled         *bool          `gorm:"column:chat_ai_enabled;default:false;comment:是否启用AI聊天功能" json:"chat_ai_enabled"`
	ChatBaseURL           *string        `gorm:"column:chat_base_url;type:varchar(255);default:'';comment:聊天AI的基础URL地址" json:"chat_base_url"`
	ChatAPIKey            *string        `gorm:"column:chat_api_key;type:varchar(255);default:'';comment:聊天AI的API密钥" json:"chat_api_key"`
	WorkflowModel         *string        `gorm:"column:workflow_model;type:varchar(100);default:'';comment:聊天AI使用的模型名称" json:"workflow_model"`
	ChatModel             *string        `gorm:"column:chat_model;type:varchar(100);default:'';comment:聊天AI使用的模型名称" json:"chat_model"`
	ImageRecognitionModel *string        `gorm:"column:image_recognition_model;type:varchar(100);default:'';comment:图像识别AI使用的模型名称" json:"image_recognition_model"`
	ChatPrompt            *string        `gorm:"column:chat_prompt;type:text;comment:聊天AI系统提示词" json:"chat_prompt"`
	MaxCompletionTokens   *int           `gorm:"column:max_completion_tokens;default:0;comment:最大回复" json:"max_completion_tokens"`
	ImageAIEnabled        *bool          `gorm:"column:image_ai_enabled;default:false;comment:是否启用AI绘图功能" json:"image_ai_enabled"`
	ImageModel            *ImageModel    `gorm:"column:image_model;type:varchar(255);default:'';comment:绘图AI模型" json:"image_model"`
	ImageAISettings       datatypes.JSON `gorm:"column:image_ai_settings;type:json;comment:绘图AI配置项" json:"image_ai_settings"`
	TTSEnabled            *bool          `gorm:"column:tts_enabled;default:false;comment:是否启用AI文本转语音功能" json:"tts_enabled"`
	TTSSettings           datatypes.JSON `gorm:"column:tts_settings;type:json;comment:文本转语音配置项" json:"tts_settings"`
	LTTSSettings          datatypes.JSON `gorm:"column:ltts_settings;type:json;comment:长文本转语音配置项" json:"ltts_settings"`
}

// TableName 设置表名
func (FriendSettings) TableName() string {
	return "friend_settings"
}
