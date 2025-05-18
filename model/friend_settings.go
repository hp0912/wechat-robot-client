package model

import (
	"gorm.io/datatypes"
)

type FriendSettings struct {
	ID              uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:公共配置表主键ID" json:"id"`
	Owner           string         `gorm:"column:owner;type:varchar(64);default:'';comment:所有者微信ID" json:"owner"`
	WeChatID        string         `gorm:"column:wechat_id;type:varchar(64);default:'';comment:好友微信ID" json:"wechat_id"`
	ChatAIEnabled   *bool          `gorm:"column:chat_ai_enabled;default:false;comment:是否启用AI聊天功能" json:"chat_ai_enabled"`
	ChatAITrigger   *string        `gorm:"column:chat_ai_trigger;type:varchar(20);default:'';comment:触发聊天AI的关键词" json:"chat_ai_trigger"`
	ChatBaseURL     string         `gorm:"column:chat_base_url;type:varchar(255);default:'';comment:聊天AI的基础URL地址" json:"chat_base_url"`
	ChatAPIKey      string         `gorm:"column:chat_api_key;type:varchar(255);default:'';comment:聊天AI的API密钥" json:"chat_api_key"`
	ChatModel       string         `gorm:"column:chat_model;type:varchar(100);default:'';comment:聊天AI使用的模型名称" json:"chat_model"`
	ChatPrompt      string         `gorm:"column:chat_prompt;type:text;comment:聊天AI系统提示词" json:"chat_prompt"`
	ImageAIEnabled  *bool          `gorm:"column:image_ai_enabled;default:false;comment:是否启用AI绘图功能" json:"image_ai_enabled"`
	ImageModel      string         `gorm:"column:image_model;type:varchar(255);default:'';comment:绘图AI模型" json:"image_model"`
	ImageAISettings datatypes.JSON `gorm:"column:image_ai_settings;type:json;comment:绘图AI配置项" json:"image_ai_settings"`
}

// TableName 设置表名
func (FriendSettings) TableName() string {
	return "friend_settings"
}
