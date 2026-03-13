package model

// ImageKnowledgeDocument 图片知识库文档
type ImageKnowledgeDocument struct {
	ID          int64  `gorm:"primarykey" json:"id"`
	Title       string `gorm:"column:title;size:255;index" json:"title"`
	Description string `gorm:"column:description;type:text" json:"description"` // 图片描述（可选）
	ImageURL    string `gorm:"column:image_url;size:512" json:"image_url"`      // 图片 URL
	Category    string `gorm:"column:category;size:128;index" json:"category"`  // 知识分类
	VectorID    string `gorm:"column:vector_id;size:128" json:"vector_id"`      // Qdrant 中的点 ID
	Enabled     bool   `gorm:"column:enabled;default:true;index" json:"enabled"`
	CreatedAt   int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at" json:"updated_at"`
}

func (ImageKnowledgeDocument) TableName() string {
	return "image_knowledge_documents"
}
