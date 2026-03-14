package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"

	"gorm.io/gorm"
)

// RAGService 检索增强生成服务
type RAGService struct {
	db          *gorm.DB
	memorySvc   *MemoryService
	vectorStore *VectorStoreService
}

// NewRAGService 创建 RAG 服务
func NewRAGService(db *gorm.DB, memorySvc *MemoryService, vectorStore *VectorStoreService) *RAGService {
	return &RAGService{
		db:          db,
		memorySvc:   memorySvc,
		vectorStore: vectorStore,
	}
}

// RetrieveContext 检索与当前对话相关的所有上下文
func (s *RAGService) RetrieveContext(ctx context.Context, contactWxID, chatRoomID, query string) *ai.RetrievedContext {
	result := &ai.RetrievedContext{}

	// 1. 获取用户相关记忆
	if s.memorySvc != nil {
		if memories, err := s.memorySvc.GetRelevantMemories(ctx, contactWxID, query, 10); err == nil {
			result.UserMemories = memories
		} else {
			log.Printf("[RAG] 获取记忆失败: %v", err)
		}

		// 2. 获取上轮对话摘要
		result.SessionSummary = s.memorySvc.GetLastSessionSummary(ctx, contactWxID, chatRoomID)
	}

	// 3. 语义搜索历史消息
	if s.vectorStore != nil {
		robot_code := vars.RobotRuntime.RobotCode
		if messages, err := s.vectorStore.SearchMessages(ctx, robot_code, query, contactWxID, chatRoomID, 5); err == nil {
			result.RelevantMessages = messages
		} else {
			log.Printf("[RAG] 搜索历史消息失败: %v", err)
		}

		// 4. 搜索知识库
		if knowledge, err := s.vectorStore.SearchKnowledge(ctx, robot_code, query, "", 3); err == nil {
			result.KnowledgeDocs = knowledge
		} else {
			log.Printf("[RAG] 搜索知识库失败: %v", err)
		}
	}

	return result
}

// BuildEnhancedPrompt 将 RAG 检索结果组装为增强的系统提示词
func (s *RAGService) BuildEnhancedPrompt(basePrompt string, retrieved *ai.RetrievedContext) string {
	if retrieved == nil {
		return basePrompt
	}

	var sb strings.Builder
	sb.WriteString(basePrompt)

	// 注入用户记忆
	if len(retrieved.UserMemories) > 0 {
		sb.WriteString("\n\n## 关于这个用户你记住的信息:\n")
		for _, m := range retrieved.UserMemories {
			fmt.Fprintf(&sb, "- [%s] %s: %s\n", m.Type, m.Key, m.Content)
		}
	}

	// 注入上轮对话摘要
	if retrieved.SessionSummary != "" {
		sb.WriteString("\n\n## 上次对话摘要:\n")
		sb.WriteString(retrieved.SessionSummary)
		sb.WriteString("\n")
	}

	// 注入相关历史消息
	if len(retrieved.RelevantMessages) > 0 {
		sb.WriteString("\n\n## 可能相关的历史对话:\n")
		for _, msg := range retrieved.RelevantMessages {
			content := msg.Payload["content"]
			if content != "" {
				fmt.Fprintf(&sb, "- %s\n", content)
			}
		}
	}

	// 注入知识库内容
	if len(retrieved.KnowledgeDocs) > 0 {
		sb.WriteString("\n\n## 参考知识:\n")
		for _, doc := range retrieved.KnowledgeDocs {
			title := doc.Payload["title"]
			content := doc.Payload["content"]
			if content != "" {
				if title != "" {
					fmt.Fprintf(&sb, "### %s\n", title)
				}
				sb.WriteString(content)
				sb.WriteString("\n\n")
			}
		}
	}

	return sb.String()
}

// IndexMessage 将单条消息加入向量索引（异步调用）
func (s *RAGService) IndexMessage(ctx context.Context, msg *model.Message) {
	if s.vectorStore == nil || msg.Content == "" || msg.Type != model.MsgTypeText {
		return
	}

	contactWxID := msg.FromWxID
	chatRoomID := ""
	if msg.IsChatRoom {
		chatRoomID = msg.FromWxID
		contactWxID = msg.SenderWxID
	}

	if _, err := s.vectorStore.IndexMessage(ctx, vars.RobotRuntime.RobotCode, msg.ID, msg.Content, contactWxID, chatRoomID, msg.SenderWxID, msg.CreatedAt); err != nil {
		log.Printf("[RAG] 向量化消息失败: %v", err)
	}
}
