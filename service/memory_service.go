package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/qdrantx"
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

// extractedMemory LLM 提取出的记忆结构
type extractedMemory struct {
	WxID             string   `json:"wx_id"`
	Category         string   `json:"category"`
	Content          string   `json:"content"`
	Importance       int      `json:"importance"`
	HappenedAt       string   `json:"happened_at,omitempty"`
	ExpireAt         string   `json:"expire_at,omitempty"`
	ReminderAt       string   `json:"reminder_at,omitempty"`
	RelationType     string   `json:"relation_type,omitempty"`
	EmotionDirection string   `json:"emotion_direction,omitempty"`
	EmotionIntensity int      `json:"emotion_intensity,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	Participants     []string `json:"participants,omitempty"`
}

const groupObservationPrefix = "[群聊观察记录]\n"

const (
	defaultMemoryCandidateLimit  = 24
	proactiveReminderWindow      = 72 * time.Hour
	proactiveReminderCooldown    = 12 * time.Hour
	recentEmotionWindow          = 30 * 24 * time.Hour
	recentRelationshipWindow     = 180 * 24 * time.Hour
	upcomingEventWindow          = 14 * 24 * time.Hour
	memorySemanticScoreThreshold = 0.4
)

var relationWeightMap = map[string]float64{
	"romantic_partner": 22,
	"spouse":           22,
	"family":           18,
	"close_friend":     16,
	"best_friend":      16,
	"mentor":           12,
	"boss":             12,
	"colleague":        8,
	"classmate":        8,
	"rival":            14,
	"conflict":         16,
}

var tagWeightMap = map[string]float64{
	model.MemoryTagImportantPerson: 14,
	model.MemoryTagRomantic:        10,
	model.MemoryTagFamily:          8,
	model.MemoryTagCloseFriend:     7,
	model.MemoryTagConflict:        10,
	model.MemoryTagFollowUp:        12,
	model.MemoryTagSocialGraph:     8,
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
			if chatRoomID != "" && msg.Role == openai.ChatMessageRoleUser && strings.HasPrefix(msg.Content, groupObservationPrefix) {
				observation := strings.TrimSpace(strings.TrimPrefix(msg.Content, groupObservationPrefix))
				if observation != "" {
					conversationText.WriteString(observation)
					conversationText.WriteString("\n")
				}
				continue
			}
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
		contextHint = fmt.Sprintf("这是微信群(%s)中的对话，当前发言者是 %s (wx_id: %s)。对话中如果出现“昵称(wx_id): 内容”的行，表示这条信息属于括号中的 wx_id，对应记忆必须归属给该成员，而不是统一归属给当前发言者", chatRoomID, senderNickname, senderWxID)
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
7. emotion: 明显的情绪记忆（生气、委屈、焦虑、开心、感动等），以及引发情绪的人或事
8. group: 仅当信息关于整个群本身（群主题、群规则、群共识），此时 wx_id 留空

关键要求：
- 每条记忆的 content 必须是**完整的自然语言陈述句**，如"张三在北京字节跳动做后端工程师"
- 如果涉及时间，设置 happened_at（事件发生时间）和 expire_at（过期时间），格式为 "2006-01-02"
- reminder_at: 如果这条记忆未来值得助手主动跟进、主动提旧事或提醒，填一个建议提醒日期，格式同样为 "2006-01-02"；否则留空
- wx_id: 记忆所属者的 wx_id。默认填 "%s"，但如果对话中某一行已经显式写成“昵称(wx_id): 内容”，则必须使用该行中的 wx_id
- 如果是群里多人之间的关系八卦、谁和谁关系好/闹矛盾，且没有单一归属者，使用 category=relation、wx_id 留空，并在 participants 中列出相关 wx_id 或昵称
- relation_type 仅在 relation 类记忆里使用，可选值：romantic_partner, spouse, family, close_friend, best_friend, colleague, classmate, boss, mentor, rival, conflict, other
- emotion_direction 仅在 emotion 类记忆里使用，可选值：positive, negative, mixed, neutral
- emotion_intensity 取 0-10，情绪越强越高
- participants 用于列出这条记忆里涉及的重要人物（优先填 wx_id，其次昵称）
- tags 只在确实明显时填写，可选值：important_person, romantic, family, close_friend, conflict, follow_up, social_graph
- importance: 1-10。恋人/配偶/直系家人/强冲突 >= 8，重要事件和明显情绪 6-9，一般偏好 5-6，临时计划 3-4
- 不要提取模糊或无实际信息量的内容（如"用户在聊天"、"用户发了消息"）
- 不要提取助手自己的信息
- 如果同一段里既有事实又有情绪，可以拆成两条记忆

输出 JSON 数组：
[{"wx_id":"xxx","category":"profile|preference|event|relation|behavior|opinion|emotion|group","content":"自然语言陈述句","importance":1-10,"happened_at":"","expire_at":"","reminder_at":"","relation_type":"","emotion_direction":"","emotion_intensity":0,"tags":[],"participants":[]}]

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

func isValidMemoryCategory(category model.MemoryCategory) bool {
	switch category {
	case model.MemoryCategoryProfile,
		model.MemoryCategoryPreference,
		model.MemoryCategoryEvent,
		model.MemoryCategoryRelation,
		model.MemoryCategoryBehavior,
		model.MemoryCategoryOpinion,
		model.MemoryCategoryEmotion,
		model.MemoryCategoryGroup:
		return true
	}
	return false
}

// isGlobalCategory 判断该类别的记忆是否应该存为全局记忆（不绑定群）
// profile/preference/relation 这类稳定的用户事实应该跨场景复用
func isGlobalCategory(category string) bool {
	switch category {
	case "profile", "preference", "relation":
		return true
	}
	return false
}

func normalizeStringList(items []string, lower bool) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if lower {
			item = key
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func normalizeRelationType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "lover", "boyfriend", "girlfriend", "romantic":
		return "romantic_partner"
	case "wife", "husband":
		return "spouse"
	case "friend":
		return "close_friend"
	case "enemy", "grudge":
		return "conflict"
	}
	return value
}

func normalizeEmotionDirection(value string) model.MemoryEmotionDirection {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case string(model.MemoryEmotionPositive):
		return model.MemoryEmotionPositive
	case string(model.MemoryEmotionNegative):
		return model.MemoryEmotionNegative
	case string(model.MemoryEmotionMixed):
		return model.MemoryEmotionMixed
	case string(model.MemoryEmotionNeutral):
		return model.MemoryEmotionNeutral
	default:
		return ""
	}
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func isGroupScopedRelation(category model.MemoryCategory, wxID, chatRoomID string) bool {
	return category == model.MemoryCategoryRelation && strings.TrimSpace(wxID) == "" && chatRoomID != ""
}

func mergeStringLists(left, right []string) []string {
	combined := make([]string, 0, len(left)+len(right))
	combined = append(combined, left...)
	combined = append(combined, right...)
	return normalizeStringList(combined, false)
}

func enrichMemoryTags(tags []string, category model.MemoryCategory, relationType string, reminderAt int64, wxID, chatRoomID string) []string {
	result := normalizeStringList(tags, true)
	switch relationType {
	case "romantic_partner", "spouse":
		result = append(result, model.MemoryTagImportantPerson, model.MemoryTagRomantic)
	case "family":
		result = append(result, model.MemoryTagImportantPerson, model.MemoryTagFamily)
	case "close_friend", "best_friend":
		result = append(result, model.MemoryTagImportantPerson, model.MemoryTagCloseFriend)
	case "conflict", "rival":
		result = append(result, model.MemoryTagConflict)
	}
	if reminderAt > 0 {
		result = append(result, model.MemoryTagFollowUp)
	}
	if category == model.MemoryCategoryRelation && wxID == "" && chatRoomID != "" {
		result = append(result, model.MemoryTagSocialGraph)
	}
	return normalizeStringList(result, true)
}

func normalizeMemorySignalFields(memory *model.Memory) {
	memory.RelationType = normalizeRelationType(memory.RelationType)
	memory.Emotion = normalizeEmotionDirection(string(memory.Emotion))
	memory.EmotionIntensity = clampInt(memory.EmotionIntensity, 0, 10)
	memory.SetTagList(enrichMemoryTags(memory.TagList(), memory.Category, memory.RelationType, memory.ReminderAt, memory.WxID, memory.ChatRoomID))
	memory.SetParticipantList(normalizeStringList(memory.ParticipantList(), false))
}

func mergeMemorySignals(existing, incoming *model.Memory) {
	if incoming.RelationType != "" && (existing.RelationType == "" || existing.RelationType == "other") {
		existing.RelationType = incoming.RelationType
	}
	if incoming.EmotionIntensity > existing.EmotionIntensity {
		existing.EmotionIntensity = incoming.EmotionIntensity
		existing.Emotion = incoming.Emotion
	} else if existing.Emotion == "" && incoming.Emotion != "" {
		existing.Emotion = incoming.Emotion
	}
	if incoming.ReminderAt > 0 && (existing.ReminderAt == 0 || incoming.ReminderAt < existing.ReminderAt) {
		existing.ReminderAt = incoming.ReminderAt
	}
	existing.SetTagList(mergeStringLists(existing.TagList(), incoming.TagList()))
	existing.SetParticipantList(mergeStringLists(existing.ParticipantList(), incoming.ParticipantList()))
	normalizeMemorySignalFields(existing)
}

func (s *MemoryService) saveExtractedMemory(ctx context.Context, defaultWxID, chatRoomID string, mem extractedMemory) {
	category := model.MemoryCategory(strings.ToLower(strings.TrimSpace(mem.Category)))
	if !isValidMemoryCategory(category) {
		log.Printf("[Memory] 跳过无效分类记忆: %s", mem.Category)
		return
	}
	content := strings.TrimSpace(mem.Content)
	if content == "" {
		return
	}
	importance := mem.Importance
	if importance < 1 {
		importance = 1
	}
	if importance > 10 {
		importance = 10
	}
	happenedAt := s.parseDateUnix(mem.HappenedAt)
	expireAt := s.parseDateUnix(mem.ExpireAt)
	reminderAt := s.parseDateUnix(mem.ReminderAt)
	relationType := normalizeRelationType(mem.RelationType)
	emotionDirection := normalizeEmotionDirection(mem.EmotionDirection)
	emotionIntensity := clampInt(mem.EmotionIntensity, 0, 10)
	rawWxID := strings.TrimSpace(mem.WxID)
	groupScopedRelation := isGroupScopedRelation(category, rawWxID, chatRoomID)
	wxID := rawWxID
	if wxID == "" && category != model.MemoryCategoryGroup && !groupScopedRelation {
		wxID = defaultWxID
	}

	// 决定记忆的实际作用域：
	// - 群级别记忆(group): wxID="", chatRoomID=群ID
	// - 群内多人关系(relation 且 wx_id 为空): wxID="", chatRoomID=群ID
	// - 全局个人事实(profile/preference/relation): wxID=用户, chatRoomID=""
	// - 群内局部记忆(event/behavior/opinion/emotion): wxID=用户, chatRoomID=群ID
	memoryChatRoomID := chatRoomID
	if category == model.MemoryCategoryGroup {
		wxID = ""
	} else if groupScopedRelation {
		wxID = ""
	} else if isGlobalCategory(string(category)) {
		// 稳定的用户事实提升为全局记忆，不绑定群
		memoryChatRoomID = ""
	}
	tags := enrichMemoryTags(mem.Tags, category, relationType, reminderAt, wxID, memoryChatRoomID)
	participants := normalizeStringList(mem.Participants, false)
	incomingMemory := &model.Memory{
		WxID:             wxID,
		ChatRoomID:       memoryChatRoomID,
		Category:         category,
		Content:          content,
		Source:           "auto",
		Importance:       importance,
		HappenedAt:       happenedAt,
		ExpireAt:         expireAt,
		ReminderAt:       reminderAt,
		RelationType:     relationType,
		Emotion:          emotionDirection,
		EmotionIntensity: emotionIntensity,
	}
	incomingMemory.SetTagList(tags)
	incomingMemory.SetParticipantList(participants)
	normalizeMemorySignalFields(incomingMemory)

	// 去重检查：在目标作用域内语义搜索
	if s.vectorStore != nil {
		duplicateID := s.findDuplicateMemory(ctx, wxID, memoryChatRoomID, content)
		if duplicateID > 0 {
			existing, err := s.memoryRepo.GetByID(duplicateID)
			if err == nil && existing != nil {
				existing.Content = content
				if importance > existing.Importance {
					existing.Importance = importance
				}
				if happenedAt > 0 {
					existing.HappenedAt = happenedAt
				}
				if expireAt > 0 {
					existing.ExpireAt = expireAt
				}
				mergeMemorySignals(existing, incomingMemory)
				if err := s.memoryRepo.Update(existing); err != nil {
					log.Printf("[Memory] 更新记忆失败: %v", err)
				}
				s.indexMemoryVector(ctx, existing)
				return
			}
		}
	}

	// 创建新记忆
	if err := s.memoryRepo.Create(incomingMemory); err != nil {
		log.Printf("[Memory] 创建记忆失败: %v", err)
		return
	}
	s.indexMemoryVector(ctx, incomingMemory)
}

func (s *MemoryService) validateManualMemory(memory *model.Memory) error {
	memory.WxID = strings.TrimSpace(memory.WxID)
	memory.ChatRoomID = strings.TrimSpace(memory.ChatRoomID)
	memory.Content = strings.TrimSpace(memory.Content)
	memory.Category = model.MemoryCategory(strings.ToLower(strings.TrimSpace(string(memory.Category))))
	normalizeMemorySignalFields(memory)
	if memory.Content == "" {
		return errors.New("记忆内容不能为空")
	}
	if !isValidMemoryCategory(memory.Category) {
		return fmt.Errorf("无效的记忆分类: %s", memory.Category)
	}
	if memory.Importance <= 0 {
		memory.Importance = 5
	}
	if memory.Importance > 10 {
		memory.Importance = 10
	}
	if memory.Category == model.MemoryCategoryGroup {
		if memory.ChatRoomID == "" {
			return errors.New("群级别记忆必须指定 chat_room_id")
		}
		if memory.WxID != "" {
			return errors.New("群级别记忆不能指定 wx_id")
		}
		return nil
	}
	if isGroupScopedRelation(memory.Category, memory.WxID, memory.ChatRoomID) {
		return nil
	}
	if memory.WxID == "" {
		return errors.New("个人记忆必须指定 wx_id")
	}
	return nil
}

// findDuplicateMemory 在精确作用域内通过语义搜索判断是否存在相似记忆（阈值 0.85）
func (s *MemoryService) findDuplicateMemory(ctx context.Context, wxID, chatRoomID, content string) int64 {
	results, err := s.vectorStore.SearchMemories(ctx, vars.RobotRuntime.RobotCode, content, wxID, chatRoomID, 3)
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
func (s *MemoryService) GetRelevantMemories(ctx context.Context, wxID, chatRoomID, query string, limit int) ([]*model.Memory, error) {
	if limit <= 0 {
		return nil, nil
	}
	candidateLimit := limit * 4
	if candidateLimit < defaultMemoryCandidateLimit {
		candidateLimit = defaultMemoryCandidateLimit
	}
	vectorScores := s.collectMemoryVectorScores(ctx, wxID, chatRoomID, query, candidateLimit)
	candidateMap := make(map[int64]*model.Memory, len(vectorScores)+candidateLimit)
	if len(vectorScores) > 0 {
		vectorIDs := make([]int64, 0, len(vectorScores))
		for id := range vectorScores {
			vectorIDs = append(vectorIDs, id)
		}
		if hitMemories, err := s.memoryRepo.GetByIDs(vectorIDs); err == nil {
			for _, m := range hitMemories {
				candidateMap[m.ID] = m
			}
		}
	}
	dbMemories, err := s.memoryRepo.ListByScope(wxID, chatRoomID, candidateLimit)
	if err != nil {
		return nil, err
	}
	for _, m := range dbMemories {
		candidateMap[m.ID] = m
	}
	if len(candidateMap) == 0 {
		return nil, nil
	}
	now := time.Now().Unix()
	result := make([]*model.Memory, 0, len(candidateMap))
	for _, m := range candidateMap {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool {
		left := memoryRankingScore(result[i], vectorScores[result[i].ID], now)
		right := memoryRankingScore(result[j], vectorScores[result[j].ID], now)
		if left == right {
			return compareMemoryTieBreak(result[i], result[j])
		}
		return left > right
	})
	if len(result) > limit {
		result = result[:limit]
	}
	s.touchMemories(result)
	return result, nil
}

// GetProactiveMemories 获取适合在当前对话中自然提起的旧事。
func (s *MemoryService) GetProactiveMemories(ctx context.Context, wxID, chatRoomID string, limit int) ([]*model.Memory, error) {
	if limit <= 0 {
		return nil, nil
	}
	now := time.Now().Unix()
	until := now + int64(proactiveReminderWindow.Seconds())
	cooldownBefore := now - int64(proactiveReminderCooldown.Seconds())
	dueMemories, err := s.memoryRepo.ListDueMemories(wxID, chatRoomID, until, cooldownBefore, limit*3)
	if err != nil {
		return nil, err
	}
	candidateLimit := limit * 5
	if candidateLimit < defaultMemoryCandidateLimit {
		candidateLimit = defaultMemoryCandidateLimit
	}
	scopeMemories, err := s.memoryRepo.ListByScope(wxID, chatRoomID, candidateLimit)
	if err != nil {
		return nil, err
	}
	candidateMap := make(map[int64]*model.Memory, len(dueMemories)+len(scopeMemories))
	dueSet := make(map[int64]bool, len(dueMemories))
	for _, m := range dueMemories {
		candidateMap[m.ID] = m
		dueSet[m.ID] = true
	}
	for _, m := range scopeMemories {
		if shouldProactivelyMentionMemory(m, now) {
			candidateMap[m.ID] = m
		}
	}
	if len(candidateMap) == 0 {
		return nil, nil
	}
	result := make([]*model.Memory, 0, len(candidateMap))
	for _, m := range candidateMap {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool {
		left := proactiveMemoryScore(result[i], dueSet[result[i].ID], now)
		right := proactiveMemoryScore(result[j], dueSet[result[j].ID], now)
		if left == right {
			return compareMemoryTieBreak(result[i], result[j])
		}
		return left > right
	})
	if len(result) > limit {
		result = result[:limit]
	}
	s.touchMemories(result)
	return result, nil
}

func (s *MemoryService) collectMemoryVectorScores(ctx context.Context, wxID, chatRoomID, query string, limit int) map[int64]float64 {
	scores := make(map[int64]float64)
	query = strings.TrimSpace(query)
	if s.vectorStore == nil || query == "" {
		return scores
	}
	rc := vars.RobotRuntime.RobotCode
	if results, err := s.vectorStore.SearchMemories(ctx, rc, query, wxID, "", limit); err == nil {
		appendMemoryVectorScores(scores, results)
	}
	if chatRoomID != "" {
		if results, err := s.vectorStore.SearchMemories(ctx, rc, query, wxID, chatRoomID, limit); err == nil {
			appendMemoryVectorScores(scores, results)
		}
		if results, err := s.vectorStore.SearchMemories(ctx, rc, query, "", chatRoomID, limit/2); err == nil {
			appendMemoryVectorScores(scores, results)
		}
	}
	return scores
}

func appendMemoryVectorScores(scoreMap map[int64]float64, results []ai.VectorSearchResult) {
	for _, result := range results {
		if float64(result.Score) < memorySemanticScoreThreshold {
			continue
		}
		id, err := strconv.ParseInt(result.Payload["memory_id"], 10, 64)
		if err != nil {
			continue
		}
		score := float64(result.Score)
		if score > scoreMap[id] {
			scoreMap[id] = score
		}
	}
}

func memoryRankingScore(memory *model.Memory, semanticScore float64, now int64) float64 {
	score := float64(memory.Importance) * 10
	score += semanticScore * 30
	score += relationWeightMap[memory.RelationType]
	for _, tag := range memory.TagList() {
		score += tagWeightMap[tag]
	}
	if memory.Category == model.MemoryCategoryEmotion {
		score += 12
	}
	if memory.EmotionIntensity > 0 {
		score += float64(memory.EmotionIntensity) * 1.8
		if memory.Emotion == model.MemoryEmotionNegative {
			score += 4
		}
	}
	score += reminderUrgencyBoost(memory, now)
	score += recencyBoost(memory, now)
	return score
}

func proactiveMemoryScore(memory *model.Memory, isDue bool, now int64) float64 {
	score := memoryRankingScore(memory, 0, now)
	if isDue {
		score += 24
	}
	if isImportantPersonMemory(memory) {
		score += 6
	}
	if isEmotionSalientMemory(memory) {
		score += 8
	}
	return score
}

func shouldProactivelyMentionMemory(memory *model.Memory, now int64) bool {
	if reminderUrgencyBoost(memory, now) >= 16 {
		return true
	}
	if memory.HasTag(model.MemoryTagFollowUp) && memory.Importance >= 6 {
		return true
	}
	if isEmotionSalientMemory(memory) {
		return true
	}
	if isImportantPersonMemory(memory) && now-memory.UpdatedAt <= int64(recentRelationshipWindow.Seconds()) {
		return true
	}
	return false
}

func isImportantPersonMemory(memory *model.Memory) bool {
	if memory.HasTag(model.MemoryTagImportantPerson) || memory.HasTag(model.MemoryTagRomantic) || memory.HasTag(model.MemoryTagFamily) {
		return true
	}
	return relationWeightMap[memory.RelationType] >= 16 && memory.Importance >= 7
}

func isEmotionSalientMemory(memory *model.Memory) bool {
	if memory.Category != model.MemoryCategoryEmotion && memory.EmotionIntensity < 7 {
		return false
	}
	if memory.UpdatedAt == 0 {
		return memory.EmotionIntensity >= 7
	}
	return time.Now().Unix()-memory.UpdatedAt <= int64(recentEmotionWindow.Seconds())
}

func reminderUrgencyBoost(memory *model.Memory, now int64) float64 {
	soon := now + int64(proactiveReminderWindow.Seconds())
	if memory.ReminderAt > 0 {
		if memory.ReminderAt <= now {
			return 28
		}
		if memory.ReminderAt <= soon {
			return 24
		}
	}
	if memory.Category == model.MemoryCategoryEvent && memory.HappenedAt > 0 {
		if memory.HappenedAt >= now-12*60*60 && memory.HappenedAt <= soon {
			return 16
		}
		if memory.HappenedAt > soon && memory.HappenedAt <= now+int64(upcomingEventWindow.Seconds()) {
			return 8
		}
	}
	return 0
}

func recencyBoost(memory *model.Memory, now int64) float64 {
	anchor := memory.UpdatedAt
	if anchor == 0 {
		anchor = memory.CreatedAt
	}
	if anchor == 0 || anchor > now {
		return 0
	}
	age := now - anchor
	switch {
	case age <= int64(7*24*time.Hour/time.Second):
		return 6
	case age <= int64(30*24*time.Hour/time.Second):
		return 4
	case age <= int64(90*24*time.Hour/time.Second):
		return 2
	default:
		return 0
	}
}

func compareMemoryTieBreak(left, right *model.Memory) bool {
	if left.Importance != right.Importance {
		return left.Importance > right.Importance
	}
	if left.UpdatedAt != right.UpdatedAt {
		return left.UpdatedAt > right.UpdatedAt
	}
	return left.ID > right.ID
}

func (s *MemoryService) touchMemories(memories []*model.Memory) {
	if len(memories) == 0 {
		return
	}
	ids := make([]int64, 0, len(memories))
	for _, m := range memories {
		ids = append(ids, m.ID)
	}
	if err := s.memoryRepo.IncrementAccessCount(ids); err != nil {
		log.Printf("[Memory] 更新记忆访问计数失败: %v", err)
	}
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
- 包含关键事实：身份、职业、偏好、重要关系、近期计划、明显情绪触发点
- 不要分类列举，要写成连贯的描述段落
- 总长度控制在200字以内
- 如果有恋人、家人、密友、冲突对象或需要后续跟进的事，要优先体现
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

// RefreshAllProfiles 批量刷新所有有记忆用户的画像（全局 + 各群）
func (s *MemoryService) RefreshAllProfiles() {
	ctx := context.Background()
	wxIDs, err := s.memoryRepo.GetDistinctWxIDs()
	if err != nil {
		log.Printf("[Memory] 获取用户列表失败: %v", err)
		return
	}
	for _, wxID := range wxIDs {
		// 刷新全局画像
		if err := s.RefreshUserProfile(ctx, wxID, ""); err != nil {
			log.Printf("[Memory] 刷新全局画像失败 (%s): %v", wxID, err)
		}
		// 每次 LLM 调用间隔 1 秒，避免触发限流
		time.Sleep(time.Second)
		// 刷新该用户有群内记忆的每个群画像
		chatRoomIDs, err := s.memoryRepo.GetDistinctChatRoomsByWxID(wxID)
		if err != nil {
			log.Printf("[Memory] 获取用户群列表失败 (%s): %v", wxID, err)
			continue
		}
		for _, chatRoomID := range chatRoomIDs {
			if err := s.RefreshUserProfile(ctx, wxID, chatRoomID); err != nil {
				log.Printf("[Memory] 刷新群画像失败 (%s in %s): %v", wxID, chatRoomID, err)
			}
			time.Sleep(time.Second)
		}
	}
}

// ── 手动记忆管理 ──────────────────────────────────────────────────────

// SaveManualMemory 手动保存记忆
func (s *MemoryService) SaveManualMemory(ctx context.Context, memory *model.Memory) error {
	if err := s.validateManualMemory(memory); err != nil {
		return err
	}
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
		if err := s.vectorStore.DeleteVectors(ctx, qdrantx.CollectionMemories, []string{memory.VectorID}); err != nil {
			return err
		}
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

// SummarizeExpiredSessions 总结过期会话（私聊和群聊使用不同的不活跃阈值）
func (s *MemoryService) SummarizeExpiredSessions(privateInactiveMinutes, groupInactiveMinutes int) {
	ctx := context.Background()
	repo := repository.NewConversationSessionRepo(ctx, s.db)

	// 分别关闭私聊和群聊的过期会话
	privateSessions, err := repo.CloseExpiredPrivateSessions(privateInactiveMinutes)
	if err != nil {
		log.Printf("[Memory] 关闭私聊过期会话失败: %v", err)
	}
	groupSessions, err := repo.CloseExpiredGroupSessions(groupInactiveMinutes)
	if err != nil {
		log.Printf("[Memory] 关闭群聊过期会话失败: %v", err)
	}

	sessions := append(privateSessions, groupSessions...)
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
		// 群聊摘要：按 chat_room_id + sender_wxid 精确过滤，避免混入其他群或私聊的消息
		messages, err = msgRepo.GetMessagesByRangeInChatRoom(session.ChatRoomID, session.FirstMsgID, session.LastMsgID, 1000, session.ContactWxID, vars.RobotRuntime.WxID)
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
