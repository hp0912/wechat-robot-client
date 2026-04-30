package ai

import (
	"context"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
)

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
	ID      string
	Score   float32
	Payload map[string]string
}

// KnowledgeService 知识库管理服务接口
type KnowledgeService interface {
	AddDocument(ctx context.Context, title, content, source, category string) error
	UpdateDocument(ctx context.Context, id int64, title, content, source string) error
	DeleteDocument(ctx context.Context, title string) error
	DeleteDocumentByID(ctx context.Context, id int64) error
	ListDocuments(ctx context.Context, category string, pager appx.Pager) ([]*model.KnowledgeDocument, int64, error)
	EnableDocument(ctx context.Context, id int64) error
	DisableDocument(ctx context.Context, id int64) error
	SearchKnowledge(ctx context.Context, query, category string, limit int) ([]VectorSearchResult, error)
	SearchKnowledgeByCategories(ctx context.Context, query string, categories []string, limit int) ([]VectorSearchResult, error)
	ReindexAll(ctx context.Context) error
}

// ImageKnowledgeService 图片知识库管理服务接口
type ImageKnowledgeService interface {
	AddImageDocument(ctx context.Context, title, description, imageURL, category string) error
	DeleteImageDocument(ctx context.Context, title string) error
	DeleteImageDocumentByID(ctx context.Context, id int64) error
	ListImageDocuments(ctx context.Context, category string, pager appx.Pager) ([]*model.ImageKnowledgeDocument, int64, error)
	SearchByText(ctx context.Context, query, category string, limit int) ([]VectorSearchResult, error)
	SearchByImage(ctx context.Context, imageURL, category string, limit int) ([]VectorSearchResult, error)
	ReindexAll(ctx context.Context) error
}
