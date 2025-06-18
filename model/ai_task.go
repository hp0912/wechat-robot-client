package model

import "gorm.io/datatypes"

type AITaskType string

const (
	AITaskTypeTTS         AITaskType = "tts"           // 长文本转语音
	AITaskTypeLongTextTTS AITaskType = "long_text_tts" // 长文本转语音
)

type AITaskStatus string

const (
	AITaskStatusPending    AITaskStatus = "pending"    // 待处理
	AITaskStatusProcessing AITaskStatus = "processing" // 处理中
	AITaskStatusCompleted  AITaskStatus = "completed"  // 已完成
	AITaskStatusFailed     AITaskStatus = "failed"     // 已失败
)

type AITask struct {
	ID               int64          `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	ContactID        string         `gorm:"column:contact_id;type:varchar(64);not null;index:idx_contact_id;comment:联系人ID，私聊时为好友微信ID，群聊时为群聊ID + 下划线 +  群聊成员微信ID" json:"contact_id"`
	MessageID        int64          `gorm:"column:message_id;not null;comment:消息ID，关联messages表的msg_id" json:"message_id"`
	AIProviderTaskID string         `gorm:"column:ai_provider_task_id;type:varchar(64);index:idx_ai_provider_task_id;comment:AI服务商任务ID" json:"ai_provider_task_id"`
	AITaskType       AITaskType     `gorm:"column:ai_task_type;type:enum('long_text_tts');not null;comment:任务类型：long_text_tts-长文本转语音" json:"ai_task_type"`
	AITaskStatus     AITaskStatus   `gorm:"column:ai_task_status;type:enum('pending','processing','completed','failed');not null;comment:任务状态：pending-待处理，processing-处理中，completed-已完成，failed-已失败" json:"ai_task_status"`
	Extra            datatypes.JSON `gorm:"column:extra;type:json;comment:额外信息" json:"extra"`
	CreatedAt        int64          `gorm:"column:created_at;not null;index:idx_created_at;comment:创建时间" json:"created_at"`
	UpdatedAt        int64          `gorm:"column:updated_at;not null;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (AITask) TableName() string {
	return "ai_task"
}
