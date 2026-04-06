package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
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

	if s.memorySvc != nil {
		// 1. 获取用户画像（始终注入）
		result.UserProfile = s.memorySvc.GetUserProfile(ctx, contactWxID, chatRoomID)

		// 2. 获取与查询相关的记忆
		if memories, err := s.memorySvc.GetRelevantMemories(ctx, contactWxID, chatRoomID, query, 10); err == nil {
			result.UserMemories = memories
		} else {
			log.Printf("[RAG] 获取记忆失败: %v", err)
		}

		// 3. 获取适合自然主动提起的旧事
		if memories, err := s.memorySvc.GetProactiveMemories(ctx, contactWxID, chatRoomID, 4); err == nil {
			result.ProactiveMemories = memories
		} else {
			log.Printf("[RAG] 获取主动记忆失败: %v", err)
		}

		// 4. 获取上轮对话摘要
		result.SessionSummary = s.memorySvc.GetLastSessionSummary(ctx, contactWxID, chatRoomID)
	}

	// 5. 语义搜索历史消息
	if s.vectorStore != nil {
		robot_code := vars.RobotRuntime.RobotCode
		if messages, err := s.vectorStore.SearchMessages(ctx, robot_code, query, contactWxID, chatRoomID, 5); err == nil {
			result.RelevantMessages = messages
		} else {
			log.Printf("[RAG] 搜索历史消息失败: %v", err)
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

	// 注入用户画像（最重要，始终在最前面）
	if retrieved.UserProfile != "" {
		sb.WriteString("\n\n## 关于当前用户的画像:\n")
		sb.WriteString(retrieved.UserProfile)
		sb.WriteString("\n")
	}

	proactiveIDs := make(map[int64]bool)
	if len(retrieved.ProactiveMemories) > 0 {
		sb.WriteString("\n\n## 适合自然主动提起的旧事:\n")
		sb.WriteString("仅在当前话题自然相关、适合表达关心或追问进展时轻量提起，不要一次性全部抛出。\n")
		for _, m := range retrieved.ProactiveMemories {
			proactiveIDs[m.ID] = true
			fmt.Fprintf(&sb, "- %s\n", formatPromptMemory(m))
		}
	}

	if len(retrieved.UserMemories) > 0 {
		sections := splitPromptMemories(retrieved.UserMemories, proactiveIDs)
		appendPromptMemorySection(&sb, "你记住的重要人物和关系", sections.ImportantPeople)
		appendPromptMemorySection(&sb, "你记住的情绪线索", sections.Emotions)
		appendPromptMemorySection(&sb, "你记住的群内关系图谱", sections.SocialGraph)
		appendPromptMemorySection(&sb, "你记住的其他相关信息", sections.Others)
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

	return sb.String()
}

type promptMemorySections struct {
	ImportantPeople []*model.Memory
	Emotions        []*model.Memory
	SocialGraph     []*model.Memory
	Others          []*model.Memory
}

func splitPromptMemories(memories []*model.Memory, skip map[int64]bool) promptMemorySections {
	var sections promptMemorySections
	for _, memory := range memories {
		if memory == nil || skip[memory.ID] {
			continue
		}
		switch {
		case isSocialGraphPromptMemory(memory):
			sections.SocialGraph = append(sections.SocialGraph, memory)
		case isImportantPersonMemory(memory):
			sections.ImportantPeople = append(sections.ImportantPeople, memory)
		case memory.Category == model.MemoryCategoryEmotion || memory.EmotionIntensity >= 6:
			sections.Emotions = append(sections.Emotions, memory)
		default:
			sections.Others = append(sections.Others, memory)
		}
	}
	return sections
}

func appendPromptMemorySection(sb *strings.Builder, title string, memories []*model.Memory) {
	if len(memories) == 0 {
		return
	}
	sb.WriteString("\n\n## ")
	sb.WriteString(title)
	sb.WriteString(":\n")
	for _, memory := range memories {
		fmt.Fprintf(sb, "- %s\n", formatPromptMemory(memory))
	}
}

func isSocialGraphPromptMemory(memory *model.Memory) bool {
	return memory.HasTag(model.MemoryTagSocialGraph) || (memory.Category == model.MemoryCategoryRelation && memory.WxID == "" && memory.ChatRoomID != "")
}

func formatPromptMemory(memory *model.Memory) string {
	labels := []string{memory.Category.DisplayName()}
	if relation := relationDisplayName(memory.RelationType); relation != "" {
		labels = append(labels, relation)
	}
	if memory.Category == model.MemoryCategoryEmotion || memory.EmotionIntensity > 0 {
		if emotion := emotionDisplayName(memory.Emotion); emotion != "" {
			labels = append(labels, fmt.Sprintf("情绪:%s/%d", emotion, memory.EmotionIntensity))
		} else {
			labels = append(labels, fmt.Sprintf("情绪强度:%d", memory.EmotionIntensity))
		}
	}
	if timeHint := memoryTimeHint(memory); timeHint != "" {
		labels = append(labels, timeHint)
	}
	if len(labels) == 0 {
		return memory.Content
	}
	return fmt.Sprintf("[%s] %s", strings.Join(labels, "/"), memory.Content)
}

func relationDisplayName(value string) string {
	switch value {
	case "romantic_partner":
		return "恋人"
	case "spouse":
		return "伴侣"
	case "family":
		return "家人"
	case "close_friend", "best_friend":
		return "密友"
	case "colleague":
		return "同事"
	case "classmate":
		return "同学"
	case "boss":
		return "上级"
	case "mentor":
		return "导师"
	case "rival":
		return "竞争关系"
	case "conflict":
		return "矛盾对象"
	case "other":
		return "关系"
	default:
		return ""
	}
}

func emotionDisplayName(value model.MemoryEmotionDirection) string {
	switch value {
	case model.MemoryEmotionPositive:
		return "正向"
	case model.MemoryEmotionNegative:
		return "负向"
	case model.MemoryEmotionMixed:
		return "复杂"
	case model.MemoryEmotionNeutral:
		return "中性"
	default:
		return ""
	}
}

func memoryTimeHint(memory *model.Memory) string {
	if memory.ReminderAt > 0 {
		return "提醒:" + time.Unix(memory.ReminderAt, 0).Format("2006-01-02")
	}
	if memory.Category == model.MemoryCategoryEvent && memory.HappenedAt > 0 {
		return "时间:" + time.Unix(memory.HappenedAt, 0).Format("2006-01-02")
	}
	return ""
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
