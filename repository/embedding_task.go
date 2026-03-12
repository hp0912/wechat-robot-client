package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type EmbeddingTask struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewEmbeddingTaskRepo(ctx context.Context, db *gorm.DB) *EmbeddingTask {
	return &EmbeddingTask{Ctx: ctx, DB: db}
}

func (r *EmbeddingTask) Create(task *model.EmbeddingTask) error {
	task.CreatedAt = time.Now().Unix()
	return r.DB.WithContext(r.Ctx).Create(task).Error
}

func (r *EmbeddingTask) BatchCreate(tasks []*model.EmbeddingTask) error {
	now := time.Now().Unix()
	for _, t := range tasks {
		t.CreatedAt = now
	}
	return r.DB.WithContext(r.Ctx).CreateInBatches(tasks, 100).Error
}

// FetchPending 获取待处理的任务
func (r *EmbeddingTask) FetchPending(limit int) ([]*model.EmbeddingTask, error) {
	var tasks []*model.EmbeddingTask
	err := r.DB.WithContext(r.Ctx).
		Where("status = ?", model.EmbeddingTaskStatusPending).
		Order("id ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// MarkProcessing 标记为处理中
func (r *EmbeddingTask) MarkProcessing(ids []int64) error {
	return r.DB.WithContext(r.Ctx).
		Model(&model.EmbeddingTask{}).
		Where("id IN ?", ids).
		Update("status", model.EmbeddingTaskStatusProcessing).Error
}

// MarkDone 标记完成
func (r *EmbeddingTask) MarkDone(id int64, vectorID string) error {
	return r.DB.WithContext(r.Ctx).
		Model(&model.EmbeddingTask{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       model.EmbeddingTaskStatusDone,
			"vector_id":    vectorID,
			"processed_at": time.Now().Unix(),
		}).Error
}

// MarkFailed 标记失败
func (r *EmbeddingTask) MarkFailed(id int64, errMsg string) error {
	return r.DB.WithContext(r.Ctx).
		Model(&model.EmbeddingTask{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       model.EmbeddingTaskStatusFailed,
			"error_msg":    errMsg,
			"processed_at": time.Now().Unix(),
		}).Error
}

// CleanupDone 清理已完成超过指定天数的任务
func (r *EmbeddingTask) CleanupDone(days int) error {
	threshold := time.Now().AddDate(0, 0, -days).Unix()
	return r.DB.WithContext(r.Ctx).
		Where("status = ? AND processed_at < ?", model.EmbeddingTaskStatusDone, threshold).
		Delete(&model.EmbeddingTask{}).Error
}

// ExistsBySourceID 检查是否已存在指定来源的任务
func (r *EmbeddingTask) ExistsBySourceID(sourceType string, sourceID int64) (bool, error) {
	var count int64
	err := r.DB.WithContext(r.Ctx).
		Model(&model.EmbeddingTask{}).
		Where("source_type = ? AND source_id = ? AND status IN ?", sourceType, sourceID,
			[]model.EmbeddingTaskStatus{model.EmbeddingTaskStatusPending, model.EmbeddingTaskStatusProcessing, model.EmbeddingTaskStatusDone}).
		Count(&count).Error
	return count > 0, err
}
