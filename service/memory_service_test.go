package service

import (
	"strings"
	"testing"
	"time"
	"wechat-robot-client/model"
)

func TestMemoryRankingScoreBoostsEmotion(t *testing.T) {
	now := time.Now().Unix()
	plain := &model.Memory{
		Category:   model.MemoryCategoryPreference,
		Importance: 6,
		UpdatedAt:  now,
	}
	emotional := &model.Memory{
		Category:         model.MemoryCategoryEmotion,
		Importance:       6,
		Emotion:          model.MemoryEmotionNegative,
		EmotionIntensity: 9,
		UpdatedAt:        now,
	}
	if memoryRankingScore(emotional, 0, now) <= memoryRankingScore(plain, 0, now) {
		t.Fatalf("expected emotional memory to rank above plain memory")
	}
}

func TestShouldProactivelyMentionMemory(t *testing.T) {
	now := time.Now().Unix()
	romantic := &model.Memory{
		Category:     model.MemoryCategoryRelation,
		Importance:   9,
		RelationType: "romantic_partner",
		UpdatedAt:    now - int64((24 * time.Hour).Seconds()),
	}
	romantic.SetTagList([]string{model.MemoryTagImportantPerson, model.MemoryTagRomantic})
	if !shouldProactivelyMentionMemory(romantic, now) {
		t.Fatalf("expected important relationship memory to be proactively mentionable")
	}

	stale := &model.Memory{
		Category:   model.MemoryCategoryPreference,
		Importance: 4,
		UpdatedAt:  now - int64((365 * 24 * time.Hour).Seconds()),
	}
	if shouldProactivelyMentionMemory(stale, now) {
		t.Fatalf("expected stale low-signal memory to be ignored")
	}
}

func TestSplitPromptMemoriesGroupsAdvancedSignals(t *testing.T) {
	important := &model.Memory{ID: 1, Category: model.MemoryCategoryRelation, Importance: 9, RelationType: "family"}
	important.SetTagList([]string{model.MemoryTagImportantPerson, model.MemoryTagFamily})
	emotion := &model.Memory{ID: 2, Category: model.MemoryCategoryEmotion, Emotion: model.MemoryEmotionNegative, EmotionIntensity: 8}
	social := &model.Memory{ID: 3, Category: model.MemoryCategoryRelation, ChatRoomID: "room@chatroom"}
	social.SetTagList([]string{model.MemoryTagSocialGraph})
	other := &model.Memory{ID: 4, Category: model.MemoryCategoryPreference}

	sections := splitPromptMemories([]*model.Memory{important, emotion, social, other}, nil)
	if len(sections.ImportantPeople) != 1 || sections.ImportantPeople[0].ID != important.ID {
		t.Fatalf("expected important relationship memory to enter ImportantPeople section")
	}
	if len(sections.Emotions) != 1 || sections.Emotions[0].ID != emotion.ID {
		t.Fatalf("expected emotion memory to enter Emotions section")
	}
	if len(sections.SocialGraph) != 1 || sections.SocialGraph[0].ID != social.ID {
		t.Fatalf("expected social graph memory to enter SocialGraph section")
	}
	if len(sections.Others) != 1 || sections.Others[0].ID != other.ID {
		t.Fatalf("expected remaining memory to enter Others section")
	}
}

func TestFormatPromptMemoryIncludesAdvancedHints(t *testing.T) {
	now := time.Now().Unix()
	memory := &model.Memory{
		Category:         model.MemoryCategoryEmotion,
		Content:          "她因为项目延期对张三很生气",
		RelationType:     "conflict",
		Emotion:          model.MemoryEmotionNegative,
		EmotionIntensity: 8,
		ReminderAt:       now + int64((24 * time.Hour).Seconds()),
	}
	formatted := formatPromptMemory(memory)
	for _, fragment := range []string{"情绪", "负向/8", "矛盾对象", "提醒:"} {
		if !strings.Contains(formatted, fragment) {
			t.Fatalf("expected formatted prompt memory to contain %q, got %q", fragment, formatted)
		}
	}
}
