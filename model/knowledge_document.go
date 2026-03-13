package model

// KnowledgeDocument 知识库文档
type KnowledgeDocument struct {
	ID         int64  `gorm:"primarykey" json:"id"`
	Title      string `gorm:"column:title;size:255;index" json:"title"`
	Content    string `gorm:"column:content;type:longtext" json:"content"`
	Source     string `gorm:"column:source;size:64" json:"source"`                                                // file / url / manual
	Category   string `gorm:"column:category;size:128;index;index:idx_chunk_category,priority:2" json:"category"` // 知识分类
	ChunkIndex int    `gorm:"column:chunk_index;index:idx_chunk_category,priority:1" json:"chunk_index"`
	ChunkTotal int    `gorm:"column:chunk_total" json:"chunk_total"`
	VectorID   string `gorm:"column:vector_id;size:128" json:"vector_id"` // Qdrant 中的点 ID
	Enabled    bool   `gorm:"column:enabled;default:true;index" json:"enabled"`
	CreatedAt  int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at" json:"updated_at"`
}

func (KnowledgeDocument) TableName() string {
	return "knowledge_documents"
}
