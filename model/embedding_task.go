package model

// EmbeddingTaskStatus 向量化任务状态
type EmbeddingTaskStatus string

const (
	EmbeddingTaskStatusPending    EmbeddingTaskStatus = "pending"
	EmbeddingTaskStatusProcessing EmbeddingTaskStatus = "processing"
	EmbeddingTaskStatusDone       EmbeddingTaskStatus = "done"
	EmbeddingTaskStatusFailed     EmbeddingTaskStatus = "failed"
)

// EmbeddingTask 向量化任务队列
type EmbeddingTask struct {
	ID          int64               `gorm:"primarykey" json:"id"`
	SourceType  string              `gorm:"column:source_type;size:32;index:idx_source,priority:1" json:"source_type"` // message / memory / knowledge
	SourceID    int64               `gorm:"column:source_id;index:idx_source,priority:2" json:"source_id"`
	Content     string              `gorm:"column:content;type:text" json:"content"`
	Status      EmbeddingTaskStatus `gorm:"column:status;size:32;default:pending;index;index:idx_source,priority:3;index:idx_status_processed,priority:1" json:"status"`
	VectorID    string              `gorm:"column:vector_id;size:128" json:"vector_id"`
	ErrorMsg    string              `gorm:"column:error_msg;type:text" json:"error_msg"`
	CreatedAt   int64               `gorm:"column:created_at" json:"created_at"`
	ProcessedAt int64               `gorm:"column:processed_at;index:idx_status_processed,priority:2" json:"processed_at"`
}

func (EmbeddingTask) TableName() string {
	return "embedding_tasks"
}
