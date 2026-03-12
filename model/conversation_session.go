package model

// ConversationSession 会话管理，用于跟踪对话轮次和生成摘要
type ConversationSession struct {
	ID           int64  `gorm:"primarykey" json:"id"`
	ContactWxID  string `gorm:"column:contact_wxid;index" json:"contact_wxid"`
	ChatRoomID   string `gorm:"column:chat_room_id;index" json:"chat_room_id"`
	Summary      string `gorm:"column:summary;type:text" json:"summary"`
	MessageCount int    `gorm:"column:message_count" json:"message_count"`
	FirstMsgID   int64  `gorm:"column:first_msg_id" json:"first_msg_id"`
	LastMsgID    int64  `gorm:"column:last_msg_id" json:"last_msg_id"`
	LastActiveAt int64  `gorm:"column:last_active_at;index" json:"last_active_at"`
	IsActive     bool   `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt    int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at" json:"updated_at"`
}

func (ConversationSession) TableName() string {
	return "conversation_sessions"
}
