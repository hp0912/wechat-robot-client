package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/openai/openai-go/v3"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	friendMemoryIntervalSeconds  = 10 * 60
	chatRoomMemoryMessageCount   = 100
	memoryExtractionMessageLimit = 300
)

type MemoryService struct {
	db          *gorm.DB
	vectorStore *VectorStoreService
	memoryRepo  *repository.Memory
	msgRepo     *repository.Message
	contactRepo *repository.Contact
	crmRepo     *repository.ChatRoomMember
	gsRepo      *repository.GlobalSettings
	crsRepo     *repository.ChatRoomSettings
	mu          sync.Mutex
}

var _ ai.MemoryService = (*MemoryService)(nil)

type extractedMemory struct {
	Scope          string   `json:"scope"`
	ContactWxID    string   `json:"contact_wxid"`
	ChatRoomID     string   `json:"chat_room_id"`
	Category       string   `json:"category"`
	Content        string   `json:"content"`
	Summary        string   `json:"summary"`
	Keywords       []string `json:"keywords"`
	Participants   []string `json:"participants"`
	Importance     int      `json:"importance"`
	Confidence     int      `json:"confidence"`
	EvidenceMsgIDs []int64  `json:"evidence_msg_ids"`
}

type extractedMemberProfile struct {
	ChatRoomID         string   `json:"chat_room_id"`
	MemberWxID         string   `json:"member_wxid"`
	Personality        string   `json:"personality"`
	Interests          []string `json:"interests"`
	CommunicationStyle string   `json:"communication_style"`
	FrequentTopics     []string `json:"frequent_topics"`
	AttitudeToBot      string   `json:"attitude_to_bot"`
	Summary            string   `json:"summary"`
	Confidence         int      `json:"confidence"`
	EvidenceMsgIDs     []int64  `json:"evidence_msg_ids"`
}

type extractedMemberRelationship struct {
	ChatRoomID     string  `json:"chat_room_id"`
	FromWxID       string  `json:"from_wxid"`
	ToWxID         string  `json:"to_wxid"`
	RelationType   string  `json:"relation_type"`
	Strength       int     `json:"strength"`
	Summary        string  `json:"summary"`
	EvidenceMsgIDs []int64 `json:"evidence_msg_ids"`
}

type memoryExtractionResult struct {
	Memories            []extractedMemory             `json:"memories"`
	MemberProfiles      []extractedMemberProfile      `json:"member_profiles"`
	MemberRelationships []extractedMemberRelationship `json:"member_relationships"`
}

func NewMemoryService(db *gorm.DB, vectorStore *VectorStoreService) *MemoryService {
	ctx := context.Background()
	return &MemoryService{
		db:          db,
		vectorStore: vectorStore,
		memoryRepo:  repository.NewMemoryRepo(ctx, db),
		msgRepo:     repository.NewMessageRepo(ctx, db),
		contactRepo: repository.NewContactRepo(ctx, db),
		crmRepo:     repository.NewChatRoomMemberRepo(ctx, db),
		gsRepo:      repository.NewGlobalSettingsRepo(ctx, db),
		crsRepo:     repository.NewChatRoomSettingsRepo(ctx, db),
	}
}

func (s *MemoryService) NotifyMessage(ctx context.Context, message *model.Message) {
	if message == nil || message.Type != model.MsgTypeText || strings.TrimSpace(message.Content) == "" {
		return
	}
	if !s.enabled() || vars.RobotRuntime.RobotCode == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if message.IsChatRoom {
		s.notifyChatRoomMessage(ctx, message)
		return
	}
	s.notifyFriendMessage(ctx, message)
}

func (s *MemoryService) enabled() bool {
	settings, err := s.gsRepo.GetGlobalSettings()
	if err != nil {
		log.Printf("[Memory] 获取全局配置失败: %v", err)
		return false
	}
	if settings == nil {
		return false
	}
	if settings.MemoryEnabled != nil && !*settings.MemoryEnabled {
		return false
	}
	return settings.ChatBaseURL != "" && settings.ChatAPIKey != "" && settings.ChatModel != "" && s.vectorStore != nil
}

func (s *MemoryService) notifyFriendMessage(ctx context.Context, message *model.Message) {
	now := time.Now().Unix()
	state, err := s.memoryRepo.GetState(vars.RobotRuntime.RobotCode, string(model.MemoryScopeFriend), message.FromWxID, "")
	if err != nil {
		log.Printf("[Memory] 获取私聊提取状态失败: %v", err)
		return
	}
	if state == nil {
		state = &model.MemoryExtractionState{
			RobotCode:        vars.RobotRuntime.RobotCode,
			Scope:            string(model.MemoryScopeFriend),
			ContactWxID:      message.FromWxID,
			WindowStartMsgID: message.ID,
			WindowStartedAt:  message.CreatedAt,
			PendingCount:     1,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := s.memoryRepo.SaveState(state); err != nil {
			log.Printf("[Memory] 创建私聊提取状态失败: %v", err)
		}
		return
	}

	if state.PendingCount == 0 || state.WindowStartMsgID == 0 {
		state.WindowStartMsgID = message.ID
		state.WindowStartedAt = message.CreatedAt
	}
	state.PendingCount++
	state.UpdatedAt = now
	if message.CreatedAt-state.WindowStartedAt < friendMemoryIntervalSeconds {
		if err := s.memoryRepo.SaveState(state); err != nil {
			log.Printf("[Memory] 更新私聊提取状态失败: %v", err)
		}
		return
	}

	messages, err := s.msgRepo.GetFriendTextMessagesInIDRange(message.FromWxID, state.WindowStartMsgID, message.ID, memoryExtractionMessageLimit)
	if err != nil {
		log.Printf("[Memory] 获取私聊消息窗口失败: %v", err)
		return
	}
	if len(messages) > 0 {
		if err := s.extractAndStore(ctx, messages, false, "", message.FromWxID); err != nil {
			log.Printf("[Memory] 私聊记忆提取失败: %v", err)
			_ = s.memoryRepo.SaveState(state)
			return
		}
	}
	state.WindowStartMsgID = message.ID + 1
	state.WindowStartedAt = message.CreatedAt
	state.PendingCount = 0
	state.LastExtractedMsgID = message.ID
	state.LastExtractedAt = now
	state.UpdatedAt = now
	if err := s.memoryRepo.SaveState(state); err != nil {
		log.Printf("[Memory] 保存私聊提取状态失败: %v", err)
	}
}

func (s *MemoryService) notifyChatRoomMessage(ctx context.Context, message *model.Message) {
	blacklist, blacklistSet, err := s.getChatRoomMemoryExtractionBlacklist(message.FromWxID)
	if err != nil {
		log.Printf("[Memory] 获取群聊记忆提取黑名单失败: %v", err)
	}
	if _, ok := blacklistSet[strings.TrimSpace(message.SenderWxID)]; ok {
		return
	}

	now := time.Now().Unix()
	state, err := s.memoryRepo.GetState(vars.RobotRuntime.RobotCode, string(model.MemoryScopeGroup), "", message.FromWxID)
	if err != nil {
		log.Printf("[Memory] 获取群聊提取状态失败: %v", err)
		return
	}
	if state == nil {
		state = &model.MemoryExtractionState{
			RobotCode:        vars.RobotRuntime.RobotCode,
			Scope:            string(model.MemoryScopeGroup),
			ChatRoomID:       message.FromWxID,
			WindowStartMsgID: message.ID,
			WindowStartedAt:  message.CreatedAt,
			PendingCount:     1,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := s.memoryRepo.SaveState(state); err != nil {
			log.Printf("[Memory] 创建群聊提取状态失败: %v", err)
		}
		return
	}

	if state.PendingCount == 0 || state.WindowStartMsgID == 0 {
		state.WindowStartMsgID = message.ID
		state.WindowStartedAt = message.CreatedAt
	}
	state.PendingCount++
	state.UpdatedAt = now
	if state.PendingCount < chatRoomMemoryMessageCount {
		if err := s.memoryRepo.SaveState(state); err != nil {
			log.Printf("[Memory] 更新群聊提取状态失败: %v", err)
		}
		return
	}

	messages, err := s.msgRepo.GetChatRoomTextMessagesInIDRangeExcludeSenders(message.FromWxID, state.WindowStartMsgID, message.ID, blacklist, memoryExtractionMessageLimit)
	if err != nil {
		log.Printf("[Memory] 获取群聊消息窗口失败: %v", err)
		return
	}
	if len(messages) > 0 {
		if err := s.extractAndStore(ctx, messages, true, message.FromWxID, ""); err != nil {
			log.Printf("[Memory] 群聊记忆提取失败: %v", err)
			_ = s.memoryRepo.SaveState(state)
			return
		}
	}
	state.WindowStartMsgID = message.ID + 1
	state.WindowStartedAt = message.CreatedAt
	state.PendingCount = 0
	state.LastExtractedMsgID = message.ID
	state.LastExtractedAt = now
	state.UpdatedAt = now
	if err := s.memoryRepo.SaveState(state); err != nil {
		log.Printf("[Memory] 保存群聊提取状态失败: %v", err)
	}
}

func (s *MemoryService) getChatRoomMemoryExtractionBlacklist(chatRoomID string) ([]string, map[string]struct{}, error) {
	settings, err := s.crsRepo.GetChatRoomSettings(chatRoomID)
	if err != nil || settings == nil {
		return nil, map[string]struct{}{}, err
	}
	wxIDs, err := settings.GetMemoryExtractionBlacklist()
	if err != nil {
		return nil, map[string]struct{}{}, err
	}
	wxIDs = normalizeStrings(wxIDs)
	set := make(map[string]struct{}, len(wxIDs))
	for _, wxID := range wxIDs {
		set[wxID] = struct{}{}
	}
	return wxIDs, set, nil
}

func (s *MemoryService) extractAndStore(ctx context.Context, messages []*model.Message, isChatRoom bool, chatRoomID, contactWxID string) error {
	settings, err := s.gsRepo.GetGlobalSettings()
	if err != nil {
		return err
	}
	if settings == nil || settings.ChatBaseURL == "" || settings.ChatAPIKey == "" || settings.ChatModel == "" {
		return nil
	}

	transcript := s.buildExtractionTranscript(messages, chatRoomID)
	if transcript == "" {
		return nil
	}
	result, err := s.extractMemoriesWithAI(ctx, settings, transcript, isChatRoom, chatRoomID, contactWxID)
	if err != nil {
		return err
	}
	for _, item := range result.Memories {
		if _, err := s.saveExtractedMemory(ctx, item, isChatRoom, chatRoomID, contactWxID); err != nil {
			log.Printf("[Memory] 保存记忆失败: %v", err)
		}
	}
	if isChatRoom {
		for _, profile := range result.MemberProfiles {
			if err := s.saveMemberProfile(profile, chatRoomID); err != nil {
				log.Printf("[Memory] 保存群成员画像失败: %v", err)
			}
		}
		for _, rel := range result.MemberRelationships {
			if err := s.saveMemberRelationship(ctx, rel, chatRoomID); err != nil {
				log.Printf("[Memory] 保存群成员关系失败: %v", err)
			}
		}
	}
	return nil
}

func (s *MemoryService) buildExtractionTranscript(messages []*model.Message, chatRoomID string) string {
	var sb strings.Builder
	mentionNormalizer := s.buildMentionNormalizer(chatRoomID)
	for _, msg := range messages {
		content, mentionedWxIDs := mentionNormalizer.Normalize(msg.Content)
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		if len([]rune(content)) > 500 {
			content = string([]rune(content)[:500]) + "..."
		}
		name := s.resolveName(chatRoomID, msg.SenderWxID)
		if len(mentionedWxIDs) > 0 {
			fmt.Fprintf(&sb, "msg_id=%d sender_wxid=%s sender_name=%s mentioned_wxids=%s created_at=%d content=%q\n", msg.ID, msg.SenderWxID, name, strings.Join(mentionedWxIDs, ","), msg.CreatedAt, content)
			continue
		}
		fmt.Fprintf(&sb, "msg_id=%d sender_wxid=%s sender_name=%s created_at=%d content=%q\n", msg.ID, msg.SenderWxID, name, msg.CreatedAt, content)
	}
	return sb.String()
}

type mentionNormalizer struct {
	replacements []mentionReplacement
}

type mentionReplacement struct {
	alias string
	wxID  string
}

func (s *MemoryService) buildMentionNormalizer(chatRoomID string) mentionNormalizer {
	if strings.TrimSpace(chatRoomID) == "" {
		return mentionNormalizer{}
	}
	members, err := s.crmRepo.GetChatRoomMembers(chatRoomID)
	if err != nil {
		log.Printf("[Memory] 获取群成员用于解析@提及失败: %v", err)
		return mentionNormalizer{}
	}
	seen := make(map[string]struct{})
	replacements := make([]mentionReplacement, 0, len(members)*3)
	for _, member := range members {
		if member == nil || strings.TrimSpace(member.WechatID) == "" {
			continue
		}
		aliases := []string{member.Remark, member.Nickname, member.Alias}
		for _, alias := range aliases {
			alias = strings.TrimSpace(alias)
			if alias == "" || strings.Contains(alias, "@") {
				continue
			}
			key := alias + "\x00" + member.WechatID
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			replacements = append(replacements, mentionReplacement{alias: alias, wxID: member.WechatID})
		}
	}
	sort.SliceStable(replacements, func(i, j int) bool {
		return len([]rune(replacements[i].alias)) > len([]rune(replacements[j].alias))
	})
	return mentionNormalizer{replacements: replacements}
}

func (n mentionNormalizer) Normalize(content string) (string, []string) {
	if len(n.replacements) == 0 || content == "" {
		return content, nil
	}
	mentionedSet := make(map[string]struct{})
	result := content
	for _, replacement := range n.replacements {
		for _, delimiter := range []string{"\u2005", " "} {
			mentionText := "@" + replacement.alias + delimiter
			if !strings.Contains(result, mentionText) {
				continue
			}
			result = strings.ReplaceAll(result, mentionText, "@"+replacement.wxID+delimiter)
			mentionedSet[replacement.wxID] = struct{}{}
		}
	}
	if len(mentionedSet) == 0 {
		return result, nil
	}
	mentionedWxIDs := make([]string, 0, len(mentionedSet))
	for wxID := range mentionedSet {
		mentionedWxIDs = append(mentionedWxIDs, wxID)
	}
	sort.Strings(mentionedWxIDs)
	return result, mentionedWxIDs
}

func (s *MemoryService) extractMemoriesWithAI(ctx context.Context, settings *model.GlobalSettings, transcript string, isChatRoom bool, chatRoomID, contactWxID string) (*memoryExtractionResult, error) {
	scene := "私聊"
	scopeRule := fmt.Sprintf("只能产生 scope=friend 的记忆，contact_wxid 必须是 %s，chat_room_id 必须为空。", contactWxID)
	if isChatRoom {
		scene = "群聊"
		scopeRule = fmt.Sprintf("可以产生 scope=group_member、group、relation 的记忆；所有 chat_room_id 必须是 %s；群成员画像和关系必须使用 sender_wxid，不要使用昵称作为主键。", chatRoomID)
	}

	systemPrompt := fmt.Sprintf(`你是微信聊天机器人的长期记忆抽取器。当前场景：%s。
目标：从聊天窗口中抽取稳定、长期有用、可被未来对话使用的记忆。
硬性规则：
1. 存储身份必须使用微信 ID 字段，不要把昵称写入 contact_wxid/member_wxid/from_wxid/to_wxid。
2. %s
3. 只记录稳定偏好、身份背景、长期兴趣、重要事实、重要事件、群成员画像、群内人际关系；不要记录普通寒暄、一次性玩笑、无意义闲聊。
4. 不确定的信息降低 confidence；没有价值时返回空数组。
5. content/summary 可以是中文自然语言，但涉及具体人时优先用微信 ID 表达，系统展示时会再转换昵称。
6. relation_type 用 friend、coworker、helper、familiar、conflict、joke_partner、mentor、other 之一。
7. transcript 中的 mentioned_wxids 和 content 里的 @wxid 表示明确提及对象，可作为群成员互动和关系判断的重要证据。
`, scene, scopeRule)

	userPrompt := "聊天窗口如下：\n" + transcript
	client := newOpenAIClient(settings.ChatAPIKey, settings.ChatBaseURL)
	msg, err := streamChatCompletionMessage(ctx, &client, openai.ChatCompletionNewParams{
		Model: settings.ChatModel,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:        "wechat_memory_extraction",
					Description: openai.String("微信聊天长期记忆、群成员画像和群成员关系抽取结果。"),
					Strict:      openai.Bool(false),
					Schema:      memoryExtractionSchema(),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	var result memoryExtractionResult
	if err := json.Unmarshal([]byte(cleanJSONContent(msg.Content)), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func memoryExtractionSchema() map[string]any {
	memoryItem := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"scope":            map[string]any{"type": "string", "enum": []string{"friend", "group_member", "group", "relation"}},
			"contact_wxid":     map[string]any{"type": "string"},
			"chat_room_id":     map[string]any{"type": "string"},
			"category":         map[string]any{"type": "string", "enum": []string{"profile", "preference", "fact", "event", "relation", "emotion", "topic", "reminder"}},
			"content":          map[string]any{"type": "string"},
			"summary":          map[string]any{"type": "string"},
			"keywords":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"participants":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"importance":       map[string]any{"type": "integer", "minimum": 1, "maximum": 10},
			"confidence":       map[string]any{"type": "integer", "minimum": 1, "maximum": 100},
			"evidence_msg_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
		},
		"required": []string{"scope", "contact_wxid", "chat_room_id", "category", "content", "summary", "keywords", "participants", "importance", "confidence", "evidence_msg_ids"},
	}
	profileItem := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"chat_room_id":        map[string]any{"type": "string"},
			"member_wxid":         map[string]any{"type": "string"},
			"personality":         map[string]any{"type": "string"},
			"interests":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"communication_style": map[string]any{"type": "string"},
			"frequent_topics":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"attitude_to_bot":     map[string]any{"type": "string"},
			"summary":             map[string]any{"type": "string"},
			"confidence":          map[string]any{"type": "integer", "minimum": 1, "maximum": 100},
			"evidence_msg_ids":    map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
		},
		"required": []string{"chat_room_id", "member_wxid", "personality", "interests", "communication_style", "frequent_topics", "attitude_to_bot", "summary", "confidence", "evidence_msg_ids"},
	}
	relationItem := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"chat_room_id":     map[string]any{"type": "string"},
			"from_wxid":        map[string]any{"type": "string"},
			"to_wxid":          map[string]any{"type": "string"},
			"relation_type":    map[string]any{"type": "string", "enum": []string{"friend", "coworker", "helper", "familiar", "conflict", "joke_partner", "mentor", "other"}},
			"strength":         map[string]any{"type": "integer", "minimum": 1, "maximum": 100},
			"summary":          map[string]any{"type": "string"},
			"evidence_msg_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
		},
		"required": []string{"chat_room_id", "from_wxid", "to_wxid", "relation_type", "strength", "summary", "evidence_msg_ids"},
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"memories":             map[string]any{"type": "array", "items": memoryItem},
			"member_profiles":      map[string]any{"type": "array", "items": profileItem},
			"member_relationships": map[string]any{"type": "array", "items": relationItem},
		},
		"required": []string{"memories", "member_profiles", "member_relationships"},
	}
}

func cleanJSONContent(content string) string {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return trimmed
}

func (s *MemoryService) saveExtractedMemory(ctx context.Context, item extractedMemory, isChatRoom bool, chatRoomID, contactWxID string) (*model.Memory, error) {
	item.Content = strings.TrimSpace(item.Content)
	if item.Content == "" {
		return nil, nil
	}
	if isChatRoom {
		item.ChatRoomID = chatRoomID
		if item.Scope == "" || item.Scope == string(model.MemoryScopeFriend) {
			item.Scope = string(model.MemoryScopeGroupMember)
		}
		if item.Scope == string(model.MemoryScopeGroup) || item.Scope == string(model.MemoryScopeRelation) {
			item.ContactWxID = ""
		}
	} else {
		item.Scope = string(model.MemoryScopeFriend)
		item.ContactWxID = contactWxID
		item.ChatRoomID = ""
		item.Participants = normalizeStrings(append(item.Participants, contactWxID))
	}
	if item.Category == "" {
		item.Category = string(model.MemoryCategoryFact)
	}
	item.Participants = normalizeStrings(item.Participants)
	if item.Scope == string(model.MemoryScopeRelation) && len(item.Participants) < 2 {
		return nil, nil
	}
	now := time.Now().Unix()
	hash := memoryHash(vars.RobotRuntime.RobotCode, item.Scope, item.ContactWxID, item.ChatRoomID, item.Category, item.Content, item.Participants)
	existing, err := s.memoryRepo.GetByHash(vars.RobotRuntime.RobotCode, hash)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		existing.LastSeenAt = now
		existing.UpdatedAt = now
		existing.Importance = maxInt(existing.Importance, clampInt(item.Importance, 1, 10))
		existing.Confidence = maxInt(existing.Confidence, clampInt(item.Confidence, 1, 100))
		existing.EvidenceMsgIDs = jsonData(item.EvidenceMsgIDs)
		if err := s.memoryRepo.UpdateMemory(existing); err != nil {
			return nil, err
		}
		return existing, s.indexMemory(ctx, existing)
	}
	memory := &model.Memory{
		RobotCode:      vars.RobotRuntime.RobotCode,
		Scope:          model.MemoryScope(item.Scope),
		ContactWxID:    item.ContactWxID,
		ChatRoomID:     item.ChatRoomID,
		Category:       model.MemoryCategory(item.Category),
		Content:        item.Content,
		Summary:        item.Summary,
		Keywords:       jsonData(item.Keywords),
		Participants:   jsonData(item.Participants),
		Importance:     clampInt(defaultInt(item.Importance, 5), 1, 10),
		Confidence:     clampInt(defaultInt(item.Confidence, 70), 1, 100),
		Source:         "chat",
		EvidenceMsgIDs: jsonData(item.EvidenceMsgIDs),
		Hash:           hash,
		OccurredAt:     now,
		LastSeenAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.memoryRepo.CreateMemory(memory); err != nil {
		return nil, err
	}
	return memory, s.indexMemory(ctx, memory)
}

func (s *MemoryService) indexMemory(ctx context.Context, memory *model.Memory) error {
	if memory == nil || s.vectorStore == nil {
		return nil
	}
	participants := string(memory.Participants)
	vectorID, err := s.vectorStore.IndexMemory(ctx, memory.RobotCode, memory.ID, memory.VectorID, string(memory.Scope), string(memory.Category), memory.Content, memory.ContactWxID, memory.ChatRoomID, participants, memory.UpdatedAt)
	if err != nil {
		return err
	}
	if vectorID != memory.VectorID {
		memory.VectorID = vectorID
		return s.memoryRepo.UpdateMemory(memory)
	}
	return nil
}

func (s *MemoryService) saveMemberProfile(item extractedMemberProfile, chatRoomID string) error {
	item.MemberWxID = strings.TrimSpace(item.MemberWxID)
	if item.MemberWxID == "" {
		return nil
	}
	if item.ChatRoomID == "" {
		item.ChatRoomID = chatRoomID
	}
	if item.ChatRoomID != chatRoomID {
		return nil
	}
	now := time.Now().Unix()
	profile := &model.MemberProfile{
		RobotCode:          vars.RobotRuntime.RobotCode,
		ChatRoomID:         chatRoomID,
		MemberWxID:         item.MemberWxID,
		Personality:        item.Personality,
		Interests:          jsonData(normalizeStrings(item.Interests)),
		CommunicationStyle: item.CommunicationStyle,
		FrequentTopics:     jsonData(normalizeStrings(item.FrequentTopics)),
		AttitudeToBot:      item.AttitudeToBot,
		Summary:            item.Summary,
		Confidence:         clampInt(defaultInt(item.Confidence, 70), 1, 100),
		EvidenceMemoryIDs:  jsonData(item.EvidenceMsgIDs),
		LastAnalyzedAt:     now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	return s.memoryRepo.UpsertMemberProfile(profile)
}

func (s *MemoryService) saveMemberRelationship(ctx context.Context, item extractedMemberRelationship, chatRoomID string) error {
	fromWxID := strings.TrimSpace(item.FromWxID)
	toWxID := strings.TrimSpace(item.ToWxID)
	if fromWxID == "" || toWxID == "" || fromWxID == toWxID {
		return nil
	}
	if item.ChatRoomID != "" && item.ChatRoomID != chatRoomID {
		return nil
	}
	if fromWxID > toWxID {
		fromWxID, toWxID = toWxID, fromWxID
	}
	if item.RelationType == "" {
		item.RelationType = "other"
	}
	now := time.Now().Unix()
	rel := &model.MemberRelationship{
		RobotCode:         vars.RobotRuntime.RobotCode,
		ChatRoomID:        chatRoomID,
		FromWxID:          fromWxID,
		ToWxID:            toWxID,
		RelationType:      item.RelationType,
		Strength:          clampInt(defaultInt(item.Strength, 50), 1, 100),
		Summary:           item.Summary,
		EvidenceMemoryIDs: jsonData(item.EvidenceMsgIDs),
		LastSeenAt:        now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.memoryRepo.UpsertMemberRelationship(rel); err != nil {
		return err
	}
	_, err := s.saveExtractedMemory(ctx, extractedMemory{
		Scope:          string(model.MemoryScopeRelation),
		ChatRoomID:     chatRoomID,
		Category:       string(model.MemoryCategoryRelation),
		Content:        item.Summary,
		Summary:        item.Summary,
		Participants:   []string{fromWxID, toWxID},
		Importance:     7,
		Confidence:     clampInt(defaultInt(item.Strength, 70), 1, 100),
		EvidenceMsgIDs: item.EvidenceMsgIDs,
	}, true, chatRoomID, "")
	return err
}

func (s *MemoryService) BuildPromptContext(ctx context.Context, query, fromWxID, senderWxID string, isChatRoom bool) string {
	if strings.TrimSpace(query) == "" || !s.enabled() || s.vectorStore == nil {
		return ""
	}
	memories := make([]*model.Memory, 0)
	seen := make(map[int64]bool)
	appendSearch := func(contactWxID, chatRoomID string, limit int) {
		results, err := s.vectorStore.SearchMemories(ctx, vars.RobotRuntime.RobotCode, query, contactWxID, chatRoomID, limit)
		if err != nil {
			log.Printf("[Memory] 记忆召回失败: %v", err)
			return
		}
		vectorIDs := make([]string, 0, len(results))
		for _, result := range results {
			if result.ID != "" {
				vectorIDs = append(vectorIDs, result.ID)
			}
		}
		items, err := s.memoryRepo.GetMemoriesByVectorIDs(vectorIDs)
		if err != nil {
			log.Printf("[Memory] 查询记忆详情失败: %v", err)
			return
		}
		for _, item := range items {
			if item != nil && !seen[item.ID] {
				seen[item.ID] = true
				memories = append(memories, item)
			}
		}
	}

	if isChatRoom {
		appendSearch(senderWxID, fromWxID, 6)
		appendSearch("", fromWxID, 6)
		relationMemories, err := s.memoryRepo.ListRelationMemories(vars.RobotRuntime.RobotCode, fromWxID, senderWxID, 5)
		if err == nil {
			for _, item := range relationMemories {
				if item != nil && !seen[item.ID] {
					seen[item.ID] = true
					memories = append(memories, item)
				}
			}
		}
	} else {
		appendSearch(fromWxID, "", 8)
	}

	return s.renderPromptContext(fromWxID, senderWxID, isChatRoom, memories)
}

func (s *MemoryService) renderPromptContext(fromWxID, senderWxID string, isChatRoom bool, memories []*model.Memory) string {
	var sb strings.Builder
	if isChatRoom {
		profile, err := s.memoryRepo.GetMemberProfile(vars.RobotRuntime.RobotCode, fromWxID, senderWxID)
		if err == nil && profile != nil {
			name := s.resolveName(fromWxID, senderWxID)
			fmt.Fprintf(&sb, "[当前群成员画像]\n关于 %s：%s\n", name, s.replaceKnownWxIDs(fromWxID, profile.Summary, []string{senderWxID}))
			if profile.Personality != "" {
				fmt.Fprintf(&sb, "性格特点：%s\n", s.replaceKnownWxIDs(fromWxID, profile.Personality, []string{senderWxID}))
			}
			if profile.CommunicationStyle != "" {
				fmt.Fprintf(&sb, "沟通风格：%s\n", s.replaceKnownWxIDs(fromWxID, profile.CommunicationStyle, []string{senderWxID}))
			}
		}
		relationships, err := s.memoryRepo.ListMemberRelationships(vars.RobotRuntime.RobotCode, fromWxID, senderWxID, 5)
		if err == nil && len(relationships) > 0 {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString("[当前群成员关系]\n")
			for _, rel := range relationships {
				participants := []string{rel.FromWxID, rel.ToWxID}
				fmt.Fprintf(&sb, "- %s 与 %s：%s\n", s.resolveName(fromWxID, rel.FromWxID), s.resolveName(fromWxID, rel.ToWxID), s.replaceKnownWxIDs(fromWxID, rel.Summary, participants))
			}
		}
	}
	if len(memories) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("[长期记忆]\n")
		for _, memory := range memories {
			participants := parseStringArray(memory.Participants)
			content := s.replaceKnownWxIDs(memory.ChatRoomID, memory.Content, append(participants, memory.ContactWxID))
			switch memory.Scope {
			case model.MemoryScopeFriend:
				fmt.Fprintf(&sb, "- 关于 %s：%s\n", s.resolveName("", memory.ContactWxID), content)
			case model.MemoryScopeGroupMember:
				fmt.Fprintf(&sb, "- 关于 %s：%s\n", s.resolveName(memory.ChatRoomID, memory.ContactWxID), content)
			case model.MemoryScopeRelation:
				fmt.Fprintf(&sb, "- 群成员关系：%s\n", content)
			default:
				fmt.Fprintf(&sb, "- 群记忆：%s\n", content)
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

func (s *MemoryService) resolveName(chatRoomID, wxID string) string {
	wxID = strings.TrimSpace(wxID)
	if wxID == "" {
		return ""
	}
	if chatRoomID != "" {
		member, err := s.crmRepo.GetChatRoomMember(chatRoomID, wxID)
		if err == nil && member != nil {
			if member.Remark != "" {
				return member.Remark
			}
			if member.Nickname != "" {
				return member.Nickname
			}
			if member.Alias != "" {
				return member.Alias
			}
		}
	}
	contact, err := s.contactRepo.GetByWechatID(wxID)
	if err == nil && contact != nil {
		if contact.Remark != "" {
			return contact.Remark
		}
		if contact.Nickname != nil && *contact.Nickname != "" {
			return *contact.Nickname
		}
		if contact.Alias != "" {
			return contact.Alias
		}
	}
	return wxID
}

func (s *MemoryService) replaceKnownWxIDs(chatRoomID, content string, wxIDs []string) string {
	result := content
	for _, wxID := range normalizeStrings(wxIDs) {
		if wxID == "" {
			continue
		}
		result = strings.ReplaceAll(result, wxID, s.resolveName(chatRoomID, wxID))
	}
	return result
}

func memoryHash(robotCode, scope, contactWxID, chatRoomID, category, content string, participants []string) string {
	parts := append([]string{robotCode, scope, contactWxID, chatRoomID, category, strings.TrimSpace(content)}, normalizeStrings(participants)...)
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func normalizeStrings(values []string) []string {
	set := make(map[string]bool)
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || set[value] {
			continue
		}
		set[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func jsonData(value any) datatypes.JSON {
	data, err := json.Marshal(value)
	if err != nil {
		return datatypes.JSON([]byte("[]"))
	}
	return datatypes.JSON(data)
}

func parseStringArray(data datatypes.JSON) []string {
	var values []string
	if len(data) == 0 {
		return values
	}
	if err := json.Unmarshal(data, &values); err != nil {
		return nil
	}
	return values
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

func defaultInt(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
