package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"

	"gorm.io/gorm"
)

type ImageKnowledgeDocument struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewImageKnowledgeDocumentRepo(ctx context.Context, db *gorm.DB) *ImageKnowledgeDocument {
	return &ImageKnowledgeDocument{Ctx: ctx, DB: db}
}

func (r *ImageKnowledgeDocument) Create(doc *model.ImageKnowledgeDocument) error {
	now := time.Now().Unix()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	return r.DB.WithContext(r.Ctx).Create(doc).Error
}

func (r *ImageKnowledgeDocument) BatchCreate(docs []*model.ImageKnowledgeDocument) error {
	now := time.Now().Unix()
	for _, doc := range docs {
		doc.CreatedAt = now
		doc.UpdatedAt = now
	}
	return r.DB.WithContext(r.Ctx).CreateInBatches(docs, 100).Error
}

func (r *ImageKnowledgeDocument) Update(doc *model.ImageKnowledgeDocument) error {
	doc.UpdatedAt = time.Now().Unix()
	return r.DB.WithContext(r.Ctx).Save(doc).Error
}

func (r *ImageKnowledgeDocument) Delete(id int64) error {
	return r.DB.WithContext(r.Ctx).Delete(&model.ImageKnowledgeDocument{}, id).Error
}

func (r *ImageKnowledgeDocument) DeleteByTitle(title string) error {
	return r.DB.WithContext(r.Ctx).Where("title = ?", title).Delete(&model.ImageKnowledgeDocument{}).Error
}

func (r *ImageKnowledgeDocument) GetByID(id int64) (*model.ImageKnowledgeDocument, error) {
	var doc model.ImageKnowledgeDocument
	err := r.DB.WithContext(r.Ctx).Where("id = ?", id).First(&doc).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &doc, err
}

func (r *ImageKnowledgeDocument) GetByTitle(title string) ([]*model.ImageKnowledgeDocument, error) {
	var docs []*model.ImageKnowledgeDocument
	err := r.DB.WithContext(r.Ctx).Where("title = ?", title).Find(&docs).Error
	return docs, err
}

// List 分页获取图片知识库文档
func (r *ImageKnowledgeDocument) List(category string, pager appx.Pager) ([]*model.ImageKnowledgeDocument, int64, error) {
	var docs []*model.ImageKnowledgeDocument
	var total int64

	query := r.DB.WithContext(r.Ctx).Model(&model.ImageKnowledgeDocument{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset(pager.OffSet).Limit(pager.PageSize).Find(&docs).Error
	return docs, total, err
}

// GetAllVectorIDs 获取某个 title 下所有的向量 ID
func (r *ImageKnowledgeDocument) GetAllVectorIDs(title string) ([]string, error) {
	var ids []string
	err := r.DB.WithContext(r.Ctx).
		Model(&model.ImageKnowledgeDocument{}).
		Where("title = ? AND vector_id != ''", title).
		Pluck("vector_id", &ids).Error
	return ids, err
}
