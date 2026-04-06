package model

import (
	"encoding/json"
	"strings"
)

// MemoryCategory 记忆分类
type MemoryCategory string

// MemoryEmotionDirection 情绪方向
type MemoryEmotionDirection string

const (
	MemoryCategoryProfile    MemoryCategory = "profile"    // 基本信息：姓名、年龄、职业、所在城市等
	MemoryCategoryPreference MemoryCategory = "preference" // 偏好习惯：喜欢/不喜欢的事物、风格偏好
	MemoryCategoryEvent      MemoryCategory = "event"      // 事件计划：带时间的重要事件或计划
	MemoryCategoryRelation   MemoryCategory = "relation"   // 人际关系：家人、朋友、同事等
	MemoryCategoryBehavior   MemoryCategory = "behavior"   // 行为模式：沟通风格、活跃时间等
	MemoryCategoryOpinion    MemoryCategory = "opinion"    // 观点态度：对特定话题的看法
	MemoryCategoryEmotion    MemoryCategory = "emotion"    // 情绪记忆：生气、委屈、开心、焦虑及其触发人/事
	MemoryCategoryGroup      MemoryCategory = "group"      // 群级别信息：群主题、群文化、群共识
)

const (
	MemoryEmotionPositive MemoryEmotionDirection = "positive"
	MemoryEmotionNegative MemoryEmotionDirection = "negative"
	MemoryEmotionMixed    MemoryEmotionDirection = "mixed"
	MemoryEmotionNeutral  MemoryEmotionDirection = "neutral"
)

const (
	MemoryTagImportantPerson = "important_person"
	MemoryTagRomantic        = "romantic"
	MemoryTagFamily          = "family"
	MemoryTagCloseFriend     = "close_friend"
	MemoryTagConflict        = "conflict"
	MemoryTagFollowUp        = "follow_up"
	MemoryTagSocialGraph     = "social_graph"
)

// DisplayName 返回分类中文名，用于提示词和后台展示。
func (c MemoryCategory) DisplayName() string {
	switch c {
	case MemoryCategoryProfile:
		return "画像"
	case MemoryCategoryPreference:
		return "偏好"
	case MemoryCategoryEvent:
		return "事件"
	case MemoryCategoryRelation:
		return "关系"
	case MemoryCategoryBehavior:
		return "行为"
	case MemoryCategoryOpinion:
		return "观点"
	case MemoryCategoryEmotion:
		return "情绪"
	case MemoryCategoryGroup:
		return "群信息"
	default:
		return string(c)
	}
}

// Memory 长期记忆
//
// 每条记忆是一个独立的自然语言陈述句，例如"张三在北京字节跳动做后端工程师"。
// 通过 WxID + ChatRoomID 确定记忆的归属和可见范围：
//   - WxID 有值, ChatRoomID 为空 → 全局个人记忆（跨所有场景可见）
//   - WxID 有值, ChatRoomID 有值 → 群内个人记忆（仅在该群上下文中可见）
//   - WxID 为空, ChatRoomID 有值 → 群级别记忆（关于群本身的信息，或群内多人关系）
type Memory struct {
	ID               int64                  `gorm:"primarykey" json:"id"`
	WxID             string                 `gorm:"column:wx_id;type:varchar(64);default:'';index:idx_memory_wx;index:idx_memory_scope" json:"wx_id"`
	ChatRoomID       string                 `gorm:"column:chat_room_id;type:varchar(64);default:'';index:idx_memory_chatroom;index:idx_memory_scope" json:"chat_room_id"`
	Category         MemoryCategory         `gorm:"column:category;size:32;index" json:"category"`
	Content          string                 `gorm:"column:content;type:text" json:"content"`
	Source           string                 `gorm:"column:source;size:16;default:auto" json:"source"` // auto / manual
	Importance       int                    `gorm:"column:importance;default:5;index" json:"importance"`
	HappenedAt       int64                  `gorm:"column:happened_at;default:0" json:"happened_at"`
	ExpireAt         int64                  `gorm:"column:expire_at;default:0;index" json:"expire_at"`
	ReminderAt       int64                  `gorm:"column:reminder_at;default:0;index" json:"reminder_at"`
	RelationType     string                 `gorm:"column:relation_type;size:32;default:'';index" json:"relation_type"`
	Emotion          MemoryEmotionDirection `gorm:"column:emotion_direction;size:16;default:'';index" json:"emotion_direction"`
	EmotionIntensity int                    `gorm:"column:emotion_intensity;default:0;index" json:"emotion_intensity"`
	Tags             string                 `gorm:"column:tags;type:text" json:"tags"`
	Participants     string                 `gorm:"column:participants;type:text" json:"participants"`
	AccessCount      int                    `gorm:"column:access_count;default:0" json:"access_count"`
	LastAccessAt     int64                  `gorm:"column:last_access_at;default:0" json:"last_access_at"`
	VectorID         string                 `gorm:"column:vector_id;size:64" json:"vector_id"`
	CreatedAt        int64                  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        int64                  `gorm:"column:updated_at" json:"updated_at"`
}

func (Memory) TableName() string {
	return "memories_v2"
}

// TagList 解析标签列表。
func (m *Memory) TagList() []string {
	return decodeJSONStringList(m.Tags)
}

// ParticipantList 解析参与者列表。
func (m *Memory) ParticipantList() []string {
	return decodeJSONStringList(m.Participants)
}

// SetTagList 保存标签列表。
func (m *Memory) SetTagList(items []string) {
	m.Tags = encodeJSONStringList(items)
}

// SetParticipantList 保存参与者列表。
func (m *Memory) SetParticipantList(items []string) {
	m.Participants = encodeJSONStringList(items)
}

// HasTag 判断是否存在指定标签。
func (m *Memory) HasTag(tag string) bool {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return false
	}
	for _, item := range m.TagList() {
		if item == tag {
			return true
		}
	}
	return false
}

func decodeJSONStringList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err == nil {
		return items
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func encodeJSONStringList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	data, err := json.Marshal(items)
	if err != nil {
		return ""
	}
	return string(data)
}

// UserProfile 用户核心画像
//
// 由 LLM 从零散记忆中定期整合而成的简洁画像摘要。
// 每次 AI 对话时始终注入上下文，让 AI 持续了解用户。
// WxID + ChatRoomID 组合唯一：
//   - ChatRoomID 为空 → 全局画像
//   - ChatRoomID 有值 → 该群内的特定画像
type UserProfile struct {
	ID         int64  `gorm:"primarykey" json:"id"`
	WxID       string `gorm:"column:wx_id;type:varchar(64);default:'';uniqueIndex:uk_profile_scope" json:"wx_id"`
	ChatRoomID string `gorm:"column:chat_room_id;type:varchar(64);uniqueIndex:uk_profile_scope;default:''" json:"chat_room_id"`
	Summary    string `gorm:"column:summary;type:text" json:"summary"`
	CreatedAt  int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at" json:"updated_at"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}
