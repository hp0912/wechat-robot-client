package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type SystemPrompt struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewSystemPromptRepo(ctx context.Context, db *gorm.DB) *SystemPrompt {
	return &SystemPrompt{Ctx: ctx, DB: db}
}

func (r *SystemPrompt) Create(prompt *model.SystemPrompt) error {
	now := time.Now().Unix()
	prompt.CreatedAt = now
	prompt.UpdatedAt = now
	return r.DB.WithContext(r.Ctx).Create(prompt).Error
}

func (r *SystemPrompt) Update(id int64, title, content string) error {
	return r.DB.WithContext(r.Ctx).
		Model(&model.SystemPrompt{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"title":      title,
			"content":    content,
			"updated_at": time.Now().Unix(),
		}).Error
}

func (r *SystemPrompt) Delete(id int64) error {
	return r.DB.WithContext(r.Ctx).Delete(&model.SystemPrompt{}, id).Error
}

func (r *SystemPrompt) GetByID(id int64) (*model.SystemPrompt, error) {
	var prompt model.SystemPrompt
	err := r.DB.WithContext(r.Ctx).Where("id = ?", id).First(&prompt).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &prompt, err
}

func (r *SystemPrompt) List(keyword string) ([]*model.SystemPrompt, error) {
	var prompts []*model.SystemPrompt
	query := r.DB.WithContext(r.Ctx).Order("id DESC")
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", likeKeyword, likeKeyword)
	}
	err := query.Find(&prompts).Error
	return prompts, err
}
