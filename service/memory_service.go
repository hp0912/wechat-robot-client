package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// MemoryService 记忆管理服务
type MemoryService struct {
	db           *gorm.DB
	memoryRepo   *repository.Memory
	sessionRepo  *repository.ConversationSession
	vectorStore  *VectorStoreService
	embeddingSvc *EmbeddingService
	aiBaseURL    string
	aiAPIKey     string
	aiModel      string
}

// NewMemoryService 创建记忆服务
func NewMemoryService(db *gorm.DB, vectorStore *VectorStoreService, embeddingSvc *EmbeddingService, aiBaseURL, aiAPIKey, aiModel string) *MemoryService {
	ctx := context.Background()
	return &MemoryService{
		db:           db,
		memoryRepo:   repository.NewMemoryRepo(ctx, db),
		sessionRepo:  repository.NewConversationSessionRepo(ctx, db),
		vectorStore:  vectorStore,
		embeddingSvc: embeddingSvc,
		aiBaseURL:    aiBaseURL,
		aiAPIKey:     aiAPIKey,
		aiModel:      aiModel,
	}
}

// extractedMemory LLM 提取出的记忆结构
type extractedMemory struct {
	Type       string `json:"type"`
	Key        string `json:"key"`
	Content    string `json:"content"`
	Importance int    `json:"importance"`
}

// ExtractMemoriesFromConversation 从对话中提取记忆（异步调用）
func (s *MemoryService) ExtractMemoriesFromConversation(contactWxID, chatRoomID string, messages []openai.ChatCompletionMessage) {
	if len(messages) == 0 {
		return
	}

	ctx := context.Background()

	// 构建对话文本
	var conversationText strings.Builder
	for _, msg := range messages {
		role := "用户"
		if msg.Role == openai.ChatMessageRoleAssistant {
			role = "助手"
		}
		if msg.Content != "" {
			fmt.Fprintf(&conversationText, "%s: %s\n", role, msg.Content)
		}
	}

	// 调用 LLM 提取记忆
	extracted, err := s.callLLMExtract(ctx, conversationText.String())
	if err != nil {
		log.Printf("[Memory] 提取记忆失败: %v", err)
		return
	}

	// 保存提取出的记忆
	for _, mem := range extracted {
		s.saveOrUpdateMemory(ctx, contactWxID, chatRoomID, mem)
	}
}

func (s *MemoryService) callLLMExtract(ctx context.Context, conversation string) ([]extractedMemory, error) {
	config := openai.DefaultConfig(s.aiAPIKey)
	config.BaseURL = utils.NormalizeAIBaseURL(s.aiBaseURL)
	client := openai.NewClientWithConfig(config)

	systemPrompt := `你是一个记忆提取助手。分析以下对话内容，提取值得长期记住的信息。

提取规则：
1. 用户的个人事实（名字、年龄、职业、所在城市等）
2. 用户的偏好和习惯（喜欢的食物、风格、时间安排等）
3. 重要事件和日期（生日、纪念日、计划等）
4. 人际关系信息（家人、朋友、同事等）

输出 JSON 数组格式，每条记忆：
[{"type": "fact|preference|event|relation", "key": "简短唯一标识", "content": "详细内容", "importance": 1到10的重要性}]

如果没有值得记住的信息，返回空数组 []。只返回 JSON，不要其他内容。`

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.aiModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: conversation},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty LLM response")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	// 处理可能被 markdown 包裹的 JSON
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var memories []extractedMemory
	if err := json.Unmarshal([]byte(content), &memories); err != nil {
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}
	return memories, nil
}

func (s *MemoryService) saveOrUpdateMemory(ctx context.Context, contactWxID, chatRoomID string, mem extractedMemory) {
	// 检查是否已存在相同 key 的记忆
	existing, err := s.memoryRepo.GetByContactAndKey(contactWxID, mem.Key)
	if err != nil {
		log.Printf("[Memory] 查询记忆失败: %v", err)
		return
	}

	if existing != nil {
		// 更新已有记忆
		existing.Content = mem.Content
		if mem.Importance > existing.Importance {
			existing.Importance = mem.Importance
		}
		if err := s.memoryRepo.Update(existing); err != nil {
			log.Printf("[Memory] 更新记忆失败: %v", err)
		}
		// 更新向量
		if s.vectorStore != nil {
			s.vectorStore.IndexMemory(ctx, vars.RobotRuntime.RobotCode, existing.ID, mem.Content, contactWxID, string(existing.Type), mem.Key)
		}
		return
	}

	// 创建新记忆
	memory := &model.Memory{
		ContactWxID: contactWxID,
		ChatRoomID:  chatRoomID,
		Type:        model.MemoryType(mem.Type),
		Key:         mem.Key,
		Content:     mem.Content,
		Source:      "auto",
		Importance:  mem.Importance,
	}
	if err := s.memoryRepo.Create(memory); err != nil {
		log.Printf("[Memory] 创建记忆失败: %v", err)
		return
	}

	// 向量化
	if s.vectorStore != nil {
		if _, err := s.vectorStore.IndexMemory(ctx, vars.RobotRuntime.RobotCode, memory.ID, mem.Content, contactWxID, string(memory.Type), mem.Key); err != nil {
			log.Printf("[Memory] 向量化记忆失败: %v", err)
		}
	}
}

// GetRelevantMemories 获取与查询相关的记忆
func (s *MemoryService) GetRelevantMemories(ctx context.Context, contactWxID, query string, limit int) ([]*model.Memory, error) {
	// 先从向量库语义搜索
	var vectorResults []ai.VectorSearchResult
	if s.vectorStore != nil {
		var err error
		vectorResults, err = s.vectorStore.SearchMemories(ctx, vars.RobotRuntime.RobotCode, query, contactWxID, limit)
		if err != nil {
			log.Printf("[Memory] 向量搜索记忆失败: %v", err)
		}
	}

	// 从 MySQL 基于重要性获取
	dbMemories, err := s.memoryRepo.GetByContact(contactWxID, limit)
	if err != nil {
		return nil, err
	}

	// 合并去重：向量搜索结果优先
	seen := make(map[string]bool)
	var result []*model.Memory

	// 向量搜索命中的，直接从 payload 构造（或从 DB 获取完整数据）
	for _, vr := range vectorResults {
		if vr.Score < 0.5 { // 相似度阈值
			continue
		}
		for _, m := range dbMemories {
			if fmt.Sprintf("%d", m.ID) == vr.Payload["memory_id"] {
				if !seen[m.Key] {
					seen[m.Key] = true
					result = append(result, m)
				}
				break
			}
		}
	}

	// 补充高重要性的 DB 记忆
	for _, m := range dbMemories {
		if len(result) >= limit {
			break
		}
		if !seen[m.Key] {
			seen[m.Key] = true
			result = append(result, m)
		}
	}

	// 更新访问计数
	if len(result) > 0 {
		ids := make([]int64, len(result))
		for i, m := range result {
			ids[i] = m.ID
		}
		s.memoryRepo.IncrementAccessCount(ids)
	}

	return result, nil
}

// GetUserProfile 获取用户画像（所有 fact 类型记忆）
func (s *MemoryService) GetUserProfile(ctx context.Context, contactWxID string) ([]*model.Memory, error) {
	return s.memoryRepo.GetByContactAndType(contactWxID, model.MemoryTypeFact, 20)
}

// SaveManualMemory 手动保存记忆
func (s *MemoryService) SaveManualMemory(ctx context.Context, memory *model.Memory) error {
	memory.Source = "manual"
	if err := s.memoryRepo.Create(memory); err != nil {
		return err
	}
	if s.vectorStore != nil {
		if _, err := s.vectorStore.IndexMemory(ctx, vars.RobotRuntime.RobotCode, memory.ID, memory.Content, memory.ContactWxID, string(memory.Type), memory.Key); err != nil {
			log.Printf("[Memory] 向量化手动记忆失败: %v", err)
		}
	}
	return nil
}

// DeleteMemory 删除记忆
func (s *MemoryService) DeleteMemory(ctx context.Context, id int64) error {
	return s.memoryRepo.Delete(id)
}

// DecayOldMemories 衰减长期未访问记忆
func (s *MemoryService) DecayOldMemories() {
	ctx := context.Background()
	repo := repository.NewMemoryRepo(ctx, s.db)
	if err := repo.DecayMemories(30); err != nil {
		log.Printf("[Memory] 衰减记忆失败: %v", err)
	}
	if err := repo.DeleteExpired(); err != nil {
		log.Printf("[Memory] 删除过期记忆失败: %v", err)
	}
}

// --- 会话管理 ---

// TouchSession 更新或创建活跃会话
func (s *MemoryService) TouchSession(ctx context.Context, contactWxID, chatRoomID string, msgID int64) {
	session, err := s.sessionRepo.GetActiveSession(contactWxID, chatRoomID)
	if err != nil {
		log.Printf("[Memory] 获取活跃会话失败: %v", err)
		return
	}

	now := time.Now().Unix()
	if session != nil {
		session.LastMsgID = msgID
		session.MessageCount++
		session.LastActiveAt = now
		s.sessionRepo.Update(session)
		return
	}

	// 创建新会话
	newSession := &model.ConversationSession{
		ContactWxID:  contactWxID,
		ChatRoomID:   chatRoomID,
		MessageCount: 1,
		FirstMsgID:   msgID,
		LastMsgID:    msgID,
		LastActiveAt: now,
		IsActive:     true,
	}
	s.sessionRepo.Create(newSession)
}

// GetLastSessionSummary 获取上一轮对话摘要
func (s *MemoryService) GetLastSessionSummary(ctx context.Context, contactWxID, chatRoomID string) string {
	summary, err := s.sessionRepo.GetLatestSummary(contactWxID, chatRoomID)
	if err != nil {
		log.Printf("[Memory] 获取会话摘要失败: %v", err)
		return ""
	}
	return summary
}

// SummarizeExpiredSessions 总结过期会话
func (s *MemoryService) SummarizeExpiredSessions(inactiveMinutes int) {
	ctx := context.Background()
	repo := repository.NewConversationSessionRepo(ctx, s.db)
	sessions, err := repo.CloseExpiredSessions(inactiveMinutes)
	if err != nil {
		log.Printf("[Memory] 关闭过期会话失败: %v", err)
		return
	}

	for _, session := range sessions {
		if session.MessageCount < 3 {
			continue
		}
		summary, err := s.generateSessionSummary(ctx, session)
		if err != nil {
			log.Printf("[Memory] 生成会话摘要失败: %v", err)
			continue
		}
		session.Summary = summary
		repo.Update(session)
	}
}

func (s *MemoryService) generateSessionSummary(ctx context.Context, session *model.ConversationSession) (string, error) {
	config := openai.DefaultConfig(s.aiAPIKey)
	config.BaseURL = s.aiBaseURL
	client := openai.NewClientWithConfig(config)

	// 获取会话内的消息
	msgRepo := repository.NewMessageRepo(ctx, s.db)
	messages, err := msgRepo.GetMessagesByRange(session.FirstMsgID, session.LastMsgID, 50)
	if err != nil || len(messages) == 0 {
		return "", fmt.Errorf("get session messages: %w", err)
	}

	var conversationText strings.Builder
	for _, msg := range messages {
		fmt.Fprintf(&conversationText, "%s: %s\n", msg.SenderWxID, msg.Content)
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.aiModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是对话摘要助手。用2-3句话概括以下对话的主要内容和结论。只输出摘要，不要其他内容。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: conversationText.String(),
			},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
