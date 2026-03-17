package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type KnowledgeCategory struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewKnowledgeCategoryRepo(ctx context.Context, db *gorm.DB) *KnowledgeCategory {
	return &KnowledgeCategory{Ctx: ctx, DB: db}
}

func (r *KnowledgeCategory) Create(category *model.KnowledgeCategory) error {
	now := time.Now().Unix()
	category.CreatedAt = now
	category.UpdatedAt = now
	return r.DB.WithContext(r.Ctx).Create(category).Error
}

func (r *KnowledgeCategory) Update(id int64, name, description string) error {
	return r.DB.WithContext(r.Ctx).
		Model(&model.KnowledgeCategory{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"name":        name,
			"description": description,
			"updated_at":  time.Now().Unix(),
		}).Error
}

func (r *KnowledgeCategory) Delete(id int64) error {
	return r.DB.WithContext(r.Ctx).Delete(&model.KnowledgeCategory{}, id).Error
}

func (r *KnowledgeCategory) GetByID(id int64) (*model.KnowledgeCategory, error) {
	var category model.KnowledgeCategory
	err := r.DB.WithContext(r.Ctx).Where("id = ?", id).First(&category).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &category, err
}

func (r *KnowledgeCategory) GetByCode(code string) (*model.KnowledgeCategory, error) {
	var category model.KnowledgeCategory
	err := r.DB.WithContext(r.Ctx).Where("code = ?", code).First(&category).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &category, err
}

func (r *KnowledgeCategory) GetByCodeAndType(code string, categoryType model.KnowledgeCategoryType) (*model.KnowledgeCategory, error) {
	var category model.KnowledgeCategory
	err := r.DB.WithContext(r.Ctx).
		Where("code = ? AND type = ?", code, categoryType).
		First(&category).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &category, err
}

func (r *KnowledgeCategory) List(categoryType model.KnowledgeCategoryType) ([]*model.KnowledgeCategory, error) {
	var categories []*model.KnowledgeCategory
	query := r.DB.WithContext(r.Ctx).Order("id ASC")
	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}
	err := query.Find(&categories).Error
	return categories, err
}

func (r *KnowledgeCategory) Count() (int64, error) {
	var count int64
	err := r.DB.WithContext(r.Ctx).Model(&model.KnowledgeCategory{}).Count(&count).Error
	return count, err
}

// FirstOrCreate 按 code 查找，不存在则创建（用于种子数据初始化）
func (r *KnowledgeCategory) FirstOrCreate(category *model.KnowledgeCategory) error {
	now := time.Now().Unix()
	return r.DB.WithContext(r.Ctx).
		Where("code = ? AND type = ?", category.Code, category.Type).
		Attrs(model.KnowledgeCategory{
			Type:        category.Type,
			Name:        category.Name,
			Description: category.Description,
			IsBuiltin:   category.IsBuiltin,
			CreatedAt:   now,
			UpdatedAt:   now,
		}).
		FirstOrCreate(category).Error
}
