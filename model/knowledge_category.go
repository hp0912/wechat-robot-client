package model

// KnowledgeCategory 知识库分类（文本知识库和图片知识库共用）
type KnowledgeCategory struct {
	ID          int64  `gorm:"primarykey" json:"id"`
	Code        string `gorm:"column:code;size:64;uniqueIndex;not null" json:"code"`
	Name        string `gorm:"column:name;size:128;not null" json:"name"`
	Description string `gorm:"column:description;size:512" json:"description"`
	IsBuiltin   bool   `gorm:"column:is_builtin;default:false" json:"is_builtin"`
	CreatedAt   int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at" json:"updated_at"`
}

func (KnowledgeCategory) TableName() string {
	return "knowledge_categories"
}
