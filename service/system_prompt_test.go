package service

import "testing"

func TestValidateSystemPromptTrimsAndRequiresTitleAndContent(t *testing.T) {
	title, content, err := validateSystemPrompt("  客服人设  ", "  你是一个客服助手。  ")
	if err != nil {
		t.Fatalf("validateSystemPrompt returned error: %v", err)
	}
	if title != "客服人设" {
		t.Fatalf("title = %q, want %q", title, "客服人设")
	}
	if content != "你是一个客服助手。" {
		t.Fatalf("content = %q, want %q", content, "你是一个客服助手。")
	}

	if _, _, err := validateSystemPrompt("   ", "有效内容"); err == nil {
		t.Fatal("validateSystemPrompt accepted empty title")
	}
	if _, _, err := validateSystemPrompt("有效标题", "   "); err == nil {
		t.Fatal("validateSystemPrompt accepted empty content")
	}
}

func TestNormalizeSystemPromptKeywordTrimsKeyword(t *testing.T) {
	keyword := normalizeSystemPromptKeyword("  客服  ")
	if keyword != "客服" {
		t.Fatalf("keyword = %q, want %q", keyword, "客服")
	}
}
