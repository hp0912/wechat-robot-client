package service

import (
	"context"
	"fmt"
	"log"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/qdrantx"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"gorm.io/gorm"
)

// ImageKnowledgeService 图片知识库管理服务
type ImageKnowledgeService struct {
	db          *gorm.DB
	docRepo     *repository.ImageKnowledgeDocument
	vectorStore *VectorStoreService
}

// NewImageKnowledgeService 创建图片知识库服务
func NewImageKnowledgeService(db *gorm.DB, vectorStore *VectorStoreService) *ImageKnowledgeService {
	return &ImageKnowledgeService{
		db:          db,
		docRepo:     repository.NewImageKnowledgeDocumentRepo(context.Background(), db),
		vectorStore: vectorStore,
	}
}

// AddImageDocument 添加图片知识库文档
func (s *ImageKnowledgeService) AddImageDocument(ctx context.Context, title, description, imageURL, category string) error {
	if imageURL == "" {
		return fmt.Errorf("image_url is required")
	}

	doc := &model.ImageKnowledgeDocument{
		Title:       title,
		Description: description,
		ImageURL:    imageURL,
		Category:    category,
		Enabled:     true,
	}

	if err := s.docRepo.Create(doc); err != nil {
		return fmt.Errorf("save image document: %w", err)
	}

	// 异步向量化
	go func() {
		bgCtx := context.Background()
		vectorID, err := s.vectorStore.IndexImageKnowledge(bgCtx, vars.RobotRuntime.RobotCode, doc.ID, doc.ImageURL, doc.Title, doc.Description, doc.Category)
		if err != nil {
			log.Printf("[ImageKnowledge] 向量化图片失败 %d: %v", doc.ID, err)
			return
		}
		doc.VectorID = vectorID
		s.docRepo.Update(doc)
	}()

	return nil
}

// DeleteImageDocument 删除图片知识库文档（按 title 删除）
func (s *ImageKnowledgeService) DeleteImageDocument(ctx context.Context, title string) error {
	vectorIDs, err := s.docRepo.GetAllVectorIDs(title)
	if err != nil {
		return fmt.Errorf("get vector ids: %w", err)
	}

	if len(vectorIDs) > 0 && s.vectorStore != nil {
		if err := s.vectorStore.DeleteVectors(ctx, qdrantx.CollectionImageKnowledge, vectorIDs); err != nil {
			log.Printf("[ImageKnowledge] 删除向量失败: %v", err)
		}
	}

	return s.docRepo.DeleteByTitle(title)
}

// DeleteImageDocumentByID 按 ID 删除单个图片文档
func (s *ImageKnowledgeService) DeleteImageDocumentByID(ctx context.Context, id int64) error {
	doc, err := s.docRepo.GetByID(id)
	if err != nil || doc == nil {
		return fmt.Errorf("image document not found")
	}
	if doc.VectorID != "" && s.vectorStore != nil {
		s.vectorStore.DeleteVectors(ctx, qdrantx.CollectionImageKnowledge, []string{doc.VectorID})
	}
	return s.docRepo.Delete(id)
}

// ListImageDocuments 分页获取图片知识库文档
func (s *ImageKnowledgeService) ListImageDocuments(ctx context.Context, category string, page, pageSize int) ([]*model.ImageKnowledgeDocument, int64, error) {
	return s.docRepo.List(category, page, pageSize)
}

// SearchByText 以文搜图
func (s *ImageKnowledgeService) SearchByText(ctx context.Context, query, category string, limit int) ([]ai.VectorSearchResult, error) {
	if s.vectorStore == nil {
		return nil, fmt.Errorf("vector store not available")
	}
	return s.vectorStore.SearchImageKnowledgeByText(ctx, vars.RobotRuntime.RobotCode, query, category, limit)
}

// SearchByImage 以图搜图
func (s *ImageKnowledgeService) SearchByImage(ctx context.Context, imageURL, category string, limit int) ([]ai.VectorSearchResult, error) {
	if s.vectorStore == nil {
		return nil, fmt.Errorf("vector store not available")
	}
	return s.vectorStore.SearchImageKnowledgeByImage(ctx, vars.RobotRuntime.RobotCode, imageURL, category, limit)
}

// ReindexAll 重建所有图片知识库向量索引
func (s *ImageKnowledgeService) ReindexAll(ctx context.Context) error {
	page := 1
	pageSize := 100
	for {
		docs, total, err := s.docRepo.List("", page, pageSize)
		if err != nil {
			return err
		}
		for _, doc := range docs {
			vectorID, err := s.vectorStore.IndexImageKnowledge(ctx, vars.RobotRuntime.RobotCode, doc.ID, doc.ImageURL, doc.Title, doc.Description, doc.Category)
			if err != nil {
				log.Printf("[ImageKnowledge] 重建索引失败 %d: %v", doc.ID, err)
				continue
			}
			doc.VectorID = vectorID
			s.docRepo.Update(doc)
		}
		if int64(page*pageSize) >= total {
			break
		}
		page++
	}
	return nil
}
