package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"gorm.io/gorm"
)

const (
	// 知识库文档分块大小（字符数）
	chunkSize    = 500
	chunkOverlap = 50
)

// KnowledgeService 知识库管理服务
type KnowledgeService struct {
	db          *gorm.DB
	docRepo     *repository.KnowledgeDocument
	vectorStore *VectorStoreService
}

// NewKnowledgeService 创建知识库服务
func NewKnowledgeService(db *gorm.DB, vectorStore *VectorStoreService) *KnowledgeService {
	return &KnowledgeService{
		db:          db,
		docRepo:     repository.NewKnowledgeDocumentRepo(context.Background(), db),
		vectorStore: vectorStore,
	}
}

// AddDocument 添加知识库文档（自动分块并向量化）
func (s *KnowledgeService) AddDocument(ctx context.Context, title, content, source, category string) error {
	// 对长文本进行分块
	chunks := splitTextIntoChunks(content, chunkSize, chunkOverlap)
	if len(chunks) == 0 {
		return fmt.Errorf("content is empty")
	}

	docs := make([]*model.KnowledgeDocument, 0, len(chunks))
	for i, chunk := range chunks {
		docs = append(docs, &model.KnowledgeDocument{
			Title:      title,
			Content:    chunk,
			Source:     source,
			Category:   category,
			ChunkIndex: i,
			ChunkTotal: len(chunks),
			Enabled:    true,
		})
	}

	if err := s.docRepo.BatchCreate(docs); err != nil {
		return fmt.Errorf("save documents: %w", err)
	}

	// 异步向量化
	go func() {
		bgCtx := context.Background()
		for _, doc := range docs {
			vectorID, err := s.vectorStore.IndexKnowledge(bgCtx, vars.RobotRuntime.RobotCode, doc.ID, doc.Content, doc.Title, doc.Category)
			if err != nil {
				log.Printf("[Knowledge] 向量化文档失败 %d: %v", doc.ID, err)
				continue
			}
			doc.VectorID = vectorID
			s.docRepo.Update(doc)
		}
	}()

	return nil
}

// DeleteDocument 删除知识库文档（按 title 删除所有 chunks）
func (s *KnowledgeService) DeleteDocument(ctx context.Context, title string) error {
	// 获取所有向量 ID
	vectorIDs, err := s.docRepo.GetAllVectorIDs(title)
	if err != nil {
		return fmt.Errorf("get vector ids: %w", err)
	}

	// 从向量库删除
	if len(vectorIDs) > 0 && s.vectorStore != nil {
		if err := s.vectorStore.DeleteVectors(ctx, "knowledge", vectorIDs); err != nil {
			log.Printf("[Knowledge] 删除向量失败: %v", err)
		}
	}

	// 从数据库删除
	return s.docRepo.DeleteByTitle(title)
}

// DeleteDocumentByID 按 ID 删除单个文档
func (s *KnowledgeService) DeleteDocumentByID(ctx context.Context, id int64) error {
	doc, err := s.docRepo.GetByID(id)
	if err != nil || doc == nil {
		return fmt.Errorf("document not found")
	}
	if doc.VectorID != "" && s.vectorStore != nil {
		s.vectorStore.DeleteVectors(ctx, "knowledge", []string{doc.VectorID})
	}
	return s.docRepo.Delete(id)
}

// ListDocuments 分页获取知识库文档
func (s *KnowledgeService) ListDocuments(ctx context.Context, category string, page, pageSize int) ([]*model.KnowledgeDocument, int64, error) {
	return s.docRepo.List(category, page, pageSize)
}

// SearchKnowledge 搜索知识库（混合检索：向量 + 关键词）
func (s *KnowledgeService) SearchKnowledge(ctx context.Context, query, category string, limit int) ([]ai.VectorSearchResult, error) {
	if s.vectorStore == nil {
		return nil, fmt.Errorf("vector store not available")
	}
	return s.vectorStore.SearchKnowledge(ctx, vars.RobotRuntime.RobotCode, query, category, limit)
}

// ReindexAll 重建所有知识库向量索引
func (s *KnowledgeService) ReindexAll(ctx context.Context) error {
	page := 1
	pageSize := 100
	for {
		docs, total, err := s.docRepo.List("", page, pageSize)
		if err != nil {
			return err
		}
		for _, doc := range docs {
			// 获取该 title 的所有 chunks
			chunks, err := s.docRepo.GetByTitle(doc.Title)
			if err != nil {
				continue
			}
			for _, chunk := range chunks {
				vectorID, err := s.vectorStore.IndexKnowledge(ctx, vars.RobotRuntime.RobotCode, chunk.ID, chunk.Content, chunk.Title, chunk.Category)
				if err != nil {
					log.Printf("[Knowledge] 重建索引失败 %d: %v", chunk.ID, err)
					continue
				}
				chunk.VectorID = vectorID
				s.docRepo.Update(chunk)
			}
		}
		if int64(page*pageSize) >= total {
			break
		}
		page++
	}
	return nil
}

// splitTextIntoChunks 将文本按固定大小分块（带重叠）
func splitTextIntoChunks(text string, size, overlap int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	runes := []rune(text)
	if len(runes) <= size {
		return []string{text}
	}

	var chunks []string
	start := 0
	for start < len(runes) {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
		start += size - overlap
	}
	return chunks
}
