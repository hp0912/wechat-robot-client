package ai

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"

	"github.com/sashabaranov/go-openai"
)

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
	ID      string
	Score   float32
	Payload map[string]string
}

// RetrievedContext RAG 检索出的上下文
type RetrievedContext struct {
	UserMemories     []*model.Memory
	SessionSummary   string
	RelevantMessages []VectorSearchResult
	KnowledgeDocs    []VectorSearchResult
}

// MemoryService 记忆管理服务接口
type MemoryService interface {
	ExtractMemoriesFromConversation(contactWxID, chatRoomID string, messages []openai.ChatCompletionMessage)
	GetRelevantMemories(ctx context.Context, contactWxID, query string, limit int) ([]*model.Memory, error)
	GetUserProfile(ctx context.Context, contactWxID string) ([]*model.Memory, error)
	SaveManualMemory(ctx context.Context, memory *model.Memory) error
	DeleteMemory(ctx context.Context, id int64) error
	GetLastSessionSummary(ctx context.Context, contactWxID, chatRoomID string) string
	TouchSession(ctx context.Context, contactWxID, chatRoomID string, msgID int64)
	DecayOldMemories()
	SummarizeExpiredSessions(inactiveMinutes int)
}

// RAGService 检索增强生成服务接口
type RAGService interface {
	RetrieveContext(ctx context.Context, contactWxID, chatRoomID, query string) *RetrievedContext
	BuildEnhancedPrompt(basePrompt string, retrieved *RetrievedContext) string
	IndexMessage(ctx context.Context, msg *model.Message)
}

// KnowledgeService 知识库管理服务接口
type KnowledgeService interface {
	AddDocument(ctx context.Context, title, content, source, category string) error
	UpdateDocument(ctx context.Context, id int64, title, content, source string) error
	DeleteDocument(ctx context.Context, title string) error
	DeleteDocumentByID(ctx context.Context, id int64) error
	ListDocuments(ctx context.Context, category string, pager appx.Pager) ([]*model.KnowledgeDocument, int64, error)
	SearchKnowledge(ctx context.Context, query, category string, limit int) ([]VectorSearchResult, error)
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
