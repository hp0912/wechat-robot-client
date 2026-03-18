package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"

	"gorm.io/gorm"
)

type KnowledgeDocument struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewKnowledgeDocumentRepo(ctx context.Context, db *gorm.DB) *KnowledgeDocument {
	return &KnowledgeDocument{Ctx: ctx, DB: db}
}

func (r *KnowledgeDocument) Create(doc *model.KnowledgeDocument) error {
	now := time.Now().Unix()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	return r.DB.WithContext(r.Ctx).Create(doc).Error
}

func (r *KnowledgeDocument) BatchCreate(docs []*model.KnowledgeDocument) error {
	now := time.Now().Unix()
	for _, doc := range docs {
		doc.CreatedAt = now
		doc.UpdatedAt = now
	}
	return r.DB.WithContext(r.Ctx).CreateInBatches(docs, 100).Error
}

func (r *KnowledgeDocument) Update(doc *model.KnowledgeDocument) error {
	doc.UpdatedAt = time.Now().Unix()
	return r.DB.WithContext(r.Ctx).Save(doc).Error
}

func (r *KnowledgeDocument) Delete(id int64) error {
	return r.DB.WithContext(r.Ctx).Delete(&model.KnowledgeDocument{}, id).Error
}

func (r *KnowledgeDocument) DeleteByTitle(title string) error {
	return r.DB.WithContext(r.Ctx).Where("title = ?", title).Delete(&model.KnowledgeDocument{}).Error
}

func (r *KnowledgeDocument) DeleteByCategory(category string) error {
	return r.DB.WithContext(r.Ctx).Where("category = ?", category).Delete(&model.KnowledgeDocument{}).Error
}

func (r *KnowledgeDocument) GetByID(id int64) (*model.KnowledgeDocument, error) {
	var doc model.KnowledgeDocument
	err := r.DB.WithContext(r.Ctx).Where("id = ?", id).First(&doc).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &doc, err
}

func (r *KnowledgeDocument) GetByTitle(title string) ([]*model.KnowledgeDocument, error) {
	var docs []*model.KnowledgeDocument
	err := r.DB.WithContext(r.Ctx).Where("title = ?", title).Find(&docs).Error
	return docs, err
}

// List 分页获取知识库文档（按 title 分组只取第一个 chunk）
func (r *KnowledgeDocument) List(category string, pager appx.Pager) ([]*model.KnowledgeDocument, int64, error) {
	var docs []*model.KnowledgeDocument
	var total int64

	query := r.DB.WithContext(r.Ctx).Model(&model.KnowledgeDocument{}).Where("chunk_index = 0")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset(pager.OffSet).Limit(pager.PageSize).Find(&docs).Error
	return docs, total, err
}

// ExistsByTitle 检查指定标题的文档是否已存在
func (r *KnowledgeDocument) ExistsByTitle(title string) (bool, error) {
	var count int64
	err := r.DB.WithContext(r.Ctx).
		Model(&model.KnowledgeDocument{}).
		Where("title = ? AND chunk_index = 0", title).
		Count(&count).Error
	return count > 0, err
}

// GetAllVectorIDs 获取某个 title 下所有的向量 ID
func (r *KnowledgeDocument) GetAllVectorIDs(title string) ([]string, error) {
	var ids []string
	err := r.DB.WithContext(r.Ctx).
		Model(&model.KnowledgeDocument{}).
		Where("title = ? AND vector_id != ''", title).
		Pluck("vector_id", &ids).Error
	return ids, err
}

// GetAllVectorIDsByCategory 获取某个 category 下所有的向量 ID
func (r *KnowledgeDocument) GetAllVectorIDsByCategory(category string) ([]string, error) {
	var ids []string
	err := r.DB.WithContext(r.Ctx).
		Model(&model.KnowledgeDocument{}).
		Where("category = ? AND vector_id != ''", category).
		Pluck("vector_id", &ids).Error
	return ids, err
}

// SearchByKeyword 关键词搜索
func (r *KnowledgeDocument) SearchByKeyword(keyword string, limit int) ([]*model.KnowledgeDocument, error) {
	var docs []*model.KnowledgeDocument
	err := r.DB.WithContext(r.Ctx).
		Where("enabled = ? AND (content LIKE ? OR title LIKE ?)", true, "%"+keyword+"%", "%"+keyword+"%").
		Order("id DESC").
		Limit(limit).
		Find(&docs).Error
	return docs, err
}
