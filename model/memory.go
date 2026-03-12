package model

// MemoryType 记忆类型
type MemoryType string

const (
	MemoryTypeFact       MemoryType = "fact"       // 用户事实（名字、职业、喜好等）
	MemoryTypePreference MemoryType = "preference" // 偏好（喜欢什么风格的回复等）
	MemoryTypeEvent      MemoryType = "event"      // 重要事件（生日、纪念日等）
	MemoryTypeRelation   MemoryType = "relation"   // 关系（和其他联系人的关系）
	MemoryTypeSummary    MemoryType = "summary"    // 对话摘要
)

// Memory 长期记忆
type Memory struct {
	ID           int64      `gorm:"primarykey" json:"id"`
	ContactWxID  string     `gorm:"column:contact_wxid;index" json:"contact_wxid"`
	ChatRoomID   string     `gorm:"column:chat_room_id;index" json:"chat_room_id"`
	Type         MemoryType `gorm:"column:type;size:32;index" json:"type"`
	Key          string     `gorm:"column:key;size:255" json:"key"`
	Content      string     `gorm:"column:content;type:text" json:"content"`
	Source       string     `gorm:"column:source;size:32;default:auto" json:"source"` // auto / manual
	Importance   int        `gorm:"column:importance;default:5" json:"importance"`    // 1-10
	AccessCount  int        `gorm:"column:access_count;default:0" json:"access_count"`
	LastAccessAt int64      `gorm:"column:last_access_at" json:"last_access_at"`
	ExpireAt     int64      `gorm:"column:expire_at;default:0" json:"expire_at"`
	CreatedAt    int64      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    int64      `gorm:"column:updated_at" json:"updated_at"`
}

func (Memory) TableName() string {
	return "memories"
}
