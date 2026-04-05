package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/utils"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// MemoryService 记忆管理服务
type MemoryService struct {
	db          *gorm.DB
	memoryRepo  *repository.Memory
	profileRepo *repository.UserProfile
	sessionRepo *repository.ConversationSession
	vectorStore *VectorStoreService
	aiBaseURL   string
	aiAPIKey    string
	aiModel     string
}

// NewMemoryService 创建记忆服务
func NewMemoryService(db *gorm.DB, vectorStore *VectorStoreService, aiBaseURL, aiAPIKey, aiModel string) *MemoryService {
	ctx := context.Background()
	return &MemoryService{
		db:          db,
		memoryRepo:  repository.NewMemoryRepo(ctx, db),
		profileRepo: repository.NewUserProfileRepo(ctx, db),
		sessionRepo: repository.NewConversationSessionRepo(ctx, db),
		vectorStore: vectorStore,
		aiBaseURL:   aiBaseURL,
		aiAPIKey:    aiAPIKey,
		aiModel:     aiModel,
	}
}

// ── 记忆提取 (Extraction) ─────────────────────────────────────────────

// extractedMemory LLM 提取出的记忆结构
type extractedMemory struct {
	WxID       string `json:"wx_id"`
	Category   string `json:"category"`
	Content    string `json:"content"`
	Importance int    `json:"importance"`
	HappenedAt string `json:"happened_at,omitempty"`
	ExpireAt   string `json:"expire_at,omitempty"`
}

// ExtractMemoriesFromConversation 从对话中提取记忆（异步调用）
// senderWxID: 当前对话发送者（私聊或群聊中的某人）
// chatRoomID: 群ID（私聊时为空）
// senderNickname: 发送者昵称（用于提取提示词）
func (s *MemoryService) ExtractMemoriesFromConversation(senderWxID, chatRoomID, senderNickname string, messages []openai.ChatCompletionMessage) {
	if len(messages) == 0 {
		return
	}

	ctx := context.Background()

	// 构建对话文本：群聊需要携带身份信息
	var conversationText strings.Builder
	for _, msg := range messages {
		role := "用户"
		if msg.Role == openai.ChatMessageRoleAssistant {
			role = "助手"
		}
		if msg.Content != "" {
			if chatRoomID != "" && msg.Role == openai.ChatMessageRoleUser && senderNickname != "" {
				// 群聊中标注发言者身份
				fmt.Fprintf(&conversationText, "%s(%s): %s\n", senderNickname, senderWxID, msg.Content)
			} else {
				fmt.Fprintf(&conversationText, "%s: %s\n", role, msg.Content)
			}
		}
	}

	// 调用 LLM 提取记忆
	extracted, err := s.callLLMExtract(ctx, senderWxID, chatRoomID, senderNickname, conversationText.String())
	if err != nil {
		log.Printf("[Memory] 提取记忆失败: %v", err)
		return
	}

	// 去重 & 存储
	for _, mem := range extracted {
		s.saveExtractedMemory(ctx, senderWxID, chatRoomID, mem)
	}
}

func (s *MemoryService) callLLMExtract(ctx context.Context, senderWxID, chatRoomID, senderNickname, conversation string) ([]extractedMemory, error) {
	config := openai.DefaultConfig(s.aiAPIKey)
	config.BaseURL = utils.NormalizeAIBaseURL(s.aiBaseURL)
	client := openai.NewClientWithConfig(config)

	contextHint := "这是一段私聊对话"
	if chatRoomID != "" {
		contextHint = fmt.Sprintf("这是微信群(%s)中的对话，当前发言者是 %s (wx_id: %s)", chatRoomID, senderNickname, senderWxID)
	}

	systemPrompt := fmt.Sprintf(`你是一个记忆提取助手。%s。
分析对话内容，提取值得长期记住的**具体**信息。

提取规则：
1. profile: 个人基本信息（姓名、年龄、职业、所在城市、公司等）
2. preference: 偏好习惯（喜欢/不喜欢的事物、饮食口味、风格偏好等）
3. event: 重要事件或计划（生日、旅行计划、会议、截止日期等）
4. relation: 人际关系（家人、朋友、同事、上下级等）
5. behavior: 行为模式（沟通风格、活跃时间段、说话习惯等）
6. opinion: 明确的观点态度（对某话题的看法）
7. group: 仅当信息关于整个群本身（群主题、群规则、群共识），此时 wx_id 留空

关键要求：
- 每条记忆的 content 必须是**完整的自然语言陈述句**，如"张三在北京字节跳动做后端工程师"
- 如果涉及时间，设置 happened_at（事件发生时间）和 expire_at（过期时间），格式为 "2006-01-02"
- wx_id: 记忆所属者的 wx_id，默认填 "%s"，如果是群级别记忆则留空
- importance: 1-10，个人身份信息>=7，兴趣偏好5-6，临时计划3-4
- 不要提取模糊或无实际信息量的内容（如"用户在聊天"、"用户发了消息"）
- 不要提取助手自己的信息

输出 JSON 数组：
[{"wx_id":"xxx","category":"profile|preference|event|relation|behavior|opinion|group","content":"自然语言陈述句","importance":1-10,"happened_at":"","expire_at":""}]

没有值得记住的信息则返回 []。只返回 JSON。`, contextHint, senderWxID)

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

func (s *MemoryService) saveExtractedMemory(ctx context.Context, defaultWxID, chatRoomID string, mem extractedMemory) {
	wxID := mem.WxID
	if wxID == "" && mem.Category != "group" {
		wxID = defaultWxID
	}

	// 去重检查：用语义搜索查找是否已有相似记忆
	if s.vectorStore != nil {
		duplicateID := s.findDuplicateMemory(ctx, wxID, chatRoomID, mem.Content)
		if duplicateID > 0 {
			// 更新已有记忆而非创建新的
			existing, err := s.memoryRepo.GetByID(duplicateID)
			if err == nil && existing != nil {
				existing.Content = mem.Content
				if mem.Importance > existing.Importance {
					existing.Importance = mem.Importance
				}
				existing.HappenedAt = s.parseDateUnix(mem.HappenedAt)
				existing.ExpireAt = s.parseDateUnix(mem.ExpireAt)
				if err := s.memoryRepo.Update(existing); err != nil {
					log.Printf("[Memory] 更新记忆失败: %v", err)
				}
				s.indexMemoryVector(ctx, existing)
				return
			}
		}
	}

	// 创建新记忆
	memory := &model.Memory{
		WxID:       wxID,
		ChatRoomID: chatRoomID,
		Category:   model.MemoryCategory(mem.Category),
		Content:    mem.Content,
		Source:     "auto",
		Importance: mem.Importance,
		HappenedAt: s.parseDateUnix(mem.HappenedAt),
		ExpireAt:   s.parseDateUnix(mem.ExpireAt),
	}
	if err := s.memoryRepo.Create(memory); err != nil {
		log.Printf("[Memory] 创建记忆失败: %v", err)
		return
	}
	s.indexMemoryVector(ctx, memory)
}

// findDuplicateMemory 通过语义搜索判断是否存在相似记忆（阈值 0.85）
func (s *MemoryService) findDuplicateMemory(ctx context.Context, wxID, chatRoomID, content string) int64 {
	results, err := s.vectorStore.SearchMemories(ctx, vars.RobotRuntime.RobotCode, content, wxID, 3)
	if err != nil {
		return 0
	}
	for _, r := range results {
		if r.Score >= 0.85 {
			if id, err := strconv.ParseInt(r.Payload["memory_id"], 10, 64); err == nil {
				return id
			}
		}
	}
	return 0
}

func (s *MemoryService) indexMemoryVector(ctx context.Context, memory *model.Memory) {
	if s.vectorStore == nil {
		return
	}
	vectorID, err := s.vectorStore.IndexMemory(ctx, vars.RobotRuntime.RobotCode, memory.ID, memory.Content, memory.WxID, string(memory.Category), memory.ChatRoomID)
	if err != nil {
		log.Printf("[Memory] 向量化记忆失败: %v", err)
		return
	}
	if vectorID != "" && memory.VectorID != vectorID {
		memory.VectorID = vectorID
		s.memoryRepo.Update(memory)
	}
}

func (s *MemoryService) parseDateUnix(dateStr string) int64 {
	if dateStr == "" {
		return 0
	}
	t, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return 0
	}
	return t.Unix()
}

// ── 记忆检索 (Retrieval) ──────────────────────────────────────────────

// GetRelevantMemories 获取与当前查询相关的记忆
// 检索范围：用户全局记忆 + 群内个人记忆 + 群级别记忆
func (s *MemoryService) GetRelevantMemories(ctx context.Context, wxID, chatRoomID, query string, limit int) ([]*model.Memory, error) {
	// 语义搜索
	var vectorHitIDs []int64
	if s.vectorStore != nil && query != "" {
		results, err := s.vectorStore.SearchMemories(ctx, vars.RobotRuntime.RobotCode, query, wxID, limit*2)
		if err != nil {
			log.Printf("[Memory] 向量搜索记忆失败: %v", err)
		} else {
			for _, r := range results {
				if r.Score < 0.4 {
					continue
				}
				if id, err := strconv.ParseInt(r.Payload["memory_id"], 10, 64); err == nil {
					vectorHitIDs = append(vectorHitIDs, id)
				}
			}
		}
	}

	// 从 DB 获取高重要性记忆
	var dbMemories []*model.Memory
	var err error
	if chatRoomID != "" {
		// 群聊：全局个人 + 群内个人 + 群级别
		globalMemories, _ := s.memoryRepo.GetByWxID(wxID, limit)
		chatRoomMemories, _ := s.memoryRepo.GetByWxIDAndChatRoom(wxID, chatRoomID, limit)
		groupMemories, _ := s.memoryRepo.GetByChatRoom(chatRoomID, limit/2)
		dbMemories = append(dbMemories, globalMemories...)
		dbMemories = append(dbMemories, chatRoomMemories...)
		dbMemories = append(dbMemories, groupMemories...)
	} else {
		dbMemories, err = s.memoryRepo.GetByWxID(wxID, limit)
		if err != nil {
			return nil, err
		}
	}

	// 合并去重：语义命中优先
	seen := make(map[int64]bool)
	var result []*model.Memory

	// 先加入语义搜索命中
	if len(vectorHitIDs) > 0 {
		hitMemories, err := s.memoryRepo.GetByIDs(vectorHitIDs)
		if err == nil {
			for _, m := range hitMemories {
				if !seen[m.ID] {
					seen[m.ID] = true
					result = append(result, m)
				}
			}
		}
	}

	// 再补充 DB 高重要性记忆
	for _, m := range dbMemories {
		if len(result) >= limit {
			break
		}
		if !seen[m.ID] {
			seen[m.ID] = true
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

// ── 用户画像 (Profile) ────────────────────────────────────────────────

// GetUserProfile 获取用户画像摘要
func (s *MemoryService) GetUserProfile(ctx context.Context, wxID, chatRoomID string) string {
	// 先查全局画像
	globalProfile, _ := s.profileRepo.GetByScope(wxID, "")
	if chatRoomID == "" {
		if globalProfile != nil {
			return globalProfile.Summary
		}
		return ""
	}

	// 群聊：组合全局画像 + 群内画像
	var sb strings.Builder
	if globalProfile != nil && globalProfile.Summary != "" {
		sb.WriteString(globalProfile.Summary)
	}
	chatRoomProfile, _ := s.profileRepo.GetByScope(wxID, chatRoomID)
	if chatRoomProfile != nil && chatRoomProfile.Summary != "" {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(chatRoomProfile.Summary)
	}
	return sb.String()
}

// RefreshUserProfile 刷新用户画像（从零散记忆整合为简洁摘要）
func (s *MemoryService) RefreshUserProfile(ctx context.Context, wxID, chatRoomID string) error {
	memories, err := s.memoryRepo.GetAllByWxIDForProfile(wxID, chatRoomID)
	if err != nil || len(memories) == 0 {
		return err
	}

	// 构建记忆清单给 LLM 整合
	var memoryList strings.Builder
	for _, m := range memories {
		scope := "全局"
		if m.ChatRoomID != "" {
			scope = fmt.Sprintf("群(%s)", m.ChatRoomID)
		}
		fmt.Fprintf(&memoryList, "- [%s][%s] %s\n", scope, m.Category, m.Content)
	}

	nickname := s.resolveContactNickname(ctx, wxID, chatRoomID)
	if nickname == "" {
		nickname = wxID
	}

	config := openai.DefaultConfig(s.aiAPIKey)
	config.BaseURL = utils.NormalizeAIBaseURL(s.aiBaseURL)
	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.aiModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: `你是一个用户画像整合助手。根据以下零散记忆信息，生成一段简洁的用户画像描述。

要求：
- 用自然语言写成，像一张人物资料卡
- 包含关键事实：身份、职业、偏好、重要关系、近期计划
- 不要分类列举，要写成连贯的描述段落
- 总长度控制在200字以内
- 如果信息太少，就写少一点，不要编造`,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("用户昵称：%s\n\n记忆信息：\n%s", nickname, memoryList.String()),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("generate profile: %w", err)
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("empty response")
	}

	summary := strings.TrimSpace(resp.Choices[0].Message.Content)
	profile := &model.UserProfile{
		WxID:       wxID,
		ChatRoomID: chatRoomID,
		Summary:    summary,
	}
	return s.profileRepo.Upsert(profile)
}

// RefreshAllProfiles 批量刷新所有有记忆用户的画像
func (s *MemoryService) RefreshAllProfiles() {
	ctx := context.Background()
	wxIDs, err := s.memoryRepo.GetDistinctWxIDs()
	if err != nil {
		log.Printf("[Memory] 获取用户列表失败: %v", err)
		return
	}
	for _, wxID := range wxIDs {
		if err := s.RefreshUserProfile(ctx, wxID, ""); err != nil {
			log.Printf("[Memory] 刷新画像失败 (%s): %v", wxID, err)
		}
	}
}

// ── 手动记忆管理 ──────────────────────────────────────────────────────

// SaveManualMemory 手动保存记忆
func (s *MemoryService) SaveManualMemory(ctx context.Context, memory *model.Memory) error {
	memory.Source = "manual"
	if err := s.memoryRepo.Create(memory); err != nil {
		return err
	}
	s.indexMemoryVector(ctx, memory)
	return nil
}

// DeleteMemory 删除记忆
func (s *MemoryService) DeleteMemory(ctx context.Context, id int64) error {
	memory, err := s.memoryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if memory != nil && memory.VectorID != "" && s.vectorStore != nil {
		s.vectorStore.DeleteVectors(ctx, "memories_v2", []string{memory.VectorID})
	}
	return s.memoryRepo.Delete(id)
}

// ── 记忆衰减 & 过期清理 ───────────────────────────────────────────────

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

// ── 会话管理 ──────────────────────────────────────────────────────────

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
	if summary == "" {
		return ""
	}
	return s.replaceWxIDsInText(ctx, contactWxID, chatRoomID, summary)
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
	config.BaseURL = utils.NormalizeAIBaseURL(s.aiBaseURL)
	client := openai.NewClientWithConfig(config)

	msgRepo := repository.NewMessageRepo(ctx, s.db)
	var messages []*model.Message
	var err error
	if session.ChatRoomID != "" {
		messages, err = msgRepo.GetMessagesByRange(session.FirstMsgID, session.LastMsgID, 1000, session.ContactWxID, vars.RobotRuntime.WxID)
	} else {
		messages, err = msgRepo.GetMessagesByRange(session.FirstMsgID, session.LastMsgID, 1000)
	}
	if err != nil || len(messages) == 0 {
		return "", fmt.Errorf("get session messages: %w", err)
	}

	var conversationText strings.Builder
	for _, msg := range messages {
		nickname := msg.SenderNickname
		if nickname == "" {
			nickname = msg.SenderWxID
		}
		fmt.Fprintf(&conversationText, "%s: %s\n", nickname, msg.Content)
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

// ── 辅助方法 ──────────────────────────────────────────────────────────

func (s *MemoryService) replaceWxIDsInText(ctx context.Context, contactWxID, chatRoomID, text string) string {
	if robotWxID := vars.RobotRuntime.WxID; robotWxID != "" && strings.Contains(text, robotWxID) {
		text = strings.ReplaceAll(text, robotWxID, "助手")
	}
	if contactWxID != "" && strings.Contains(text, contactWxID) {
		if nickname := s.resolveContactNickname(ctx, contactWxID, chatRoomID); nickname != "" {
			text = strings.ReplaceAll(text, contactWxID, nickname)
		}
	}
	return text
}

func (s *MemoryService) resolveContactNickname(ctx context.Context, contactWxID, chatRoomID string) string {
	if chatRoomID != "" {
		memberRepo := repository.NewChatRoomMemberRepo(ctx, s.db)
		member, err := memberRepo.GetChatRoomMember(chatRoomID, contactWxID)
		if err == nil && member != nil {
			if member.Remark != "" {
				return member.Remark
			}
			return member.Nickname
		}
	} else {
		contactRepo := repository.NewContactRepo(ctx, s.db)
		contact, err := contactRepo.GetContact(contactWxID)
		if err == nil && contact != nil {
			if contact.Remark != "" {
				return contact.Remark
			}
			if contact.Nickname != nil {
				return *contact.Nickname
			}
		}
	}
	return ""
}

// SearchMemoriesByKeyword 关键词搜索（用于管理后台）
func (s *MemoryService) SearchMemoriesByKeyword(ctx context.Context, wxID, chatRoomID, keyword string, limit int) ([]*model.Memory, error) {
	return s.memoryRepo.SearchByKeyword(wxID, chatRoomID, keyword, limit)
}
