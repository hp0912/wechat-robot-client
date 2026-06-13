package model

// SystemPrompt 系统提示词
type SystemPrompt struct {
	ID        int64  `gorm:"primarykey" json:"id"`
	Title     string `gorm:"column:title;size:128;not null;comment:标题" json:"title"`
	Content   string `gorm:"column:content;type:text;not null;comment:提示词内容" json:"content"`
	CreatedAt int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt int64  `gorm:"column:updated_at" json:"updated_at"`
}

func (SystemPrompt) TableName() string {
	return "system_prompts"
}
