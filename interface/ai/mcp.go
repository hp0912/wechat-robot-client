package ai

import (
	"github.com/sashabaranov/go-openai"
)

// ExtraTool 额外的内置工具（由调用方注入到 ChatWithMCPTools 中）
type ExtraTool struct {
	Tool    openai.Tool
	Handler func(toolCall openai.ToolCall) (result string, immediately bool, err error)
}
