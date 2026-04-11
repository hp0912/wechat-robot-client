package openaitools

import (
	"context"
	"fmt"
	"strings"
	"wechat-robot-client/pkg/robotctx"

	"gorm.io/gorm"

	"github.com/sashabaranov/go-openai"
)

type OpenAITool interface {
	GetOpenAITool() openai.Tool
	BuildSystemPrompt(ctx context.Context, robotCtx robotctx.RobotContext) (string, error)
	ExecuteToolCall(ctx context.Context, robotCtx robotctx.RobotContext, toolCall openai.ToolCall) (string, bool, error)
}

type OpenAIToolsManager struct {
	db    *gorm.DB
	tools map[string]OpenAITool
}

// NewOpenAIToolsManager 创建 OpenAITools 管理器
func NewOpenAIToolsManager(db *gorm.DB) *OpenAIToolsManager {
	return &OpenAIToolsManager{
		db:    db,
		tools: make(map[string]OpenAITool),
	}
}

func (m *OpenAIToolsManager) Initialize() error {
	m.tools["search_knowledge"] = NewSearchKnowledgeTool()
	return nil
}

func (m *OpenAIToolsManager) Shutdown() error {
	// 如果有需要清理的资源，可以在这里处理
	return nil
}

func (m *OpenAIToolsManager) GetOpenAITools() []openai.Tool {
	var openAITools []openai.Tool
	for _, tool := range m.tools {
		openAITools = append(openAITools, tool.GetOpenAITool())
	}
	return openAITools
}

func (m *OpenAIToolsManager) IsOpenAITool(fnName string) bool {
	_, exists := m.tools[fnName]
	return exists
}

func (m *OpenAIToolsManager) BuildSystemPrompt(ctx context.Context, robotCtx robotctx.RobotContext) (string, error) {
	var sb strings.Builder
	for _, tool := range m.tools {
		prompt, err := tool.BuildSystemPrompt(ctx, robotCtx)
		if err != nil {
			return "", err
		}
		sb.WriteString("\n\n")
		sb.WriteString(prompt)
	}
	return sb.String(), nil
}

func (m *OpenAIToolsManager) ExecuteToolCall(ctx context.Context, robotCtx robotctx.RobotContext, toolCall openai.ToolCall) (string, bool, error) {
	tool, ok := m.tools[toolCall.Function.Name]
	if !ok {
		return "", false, fmt.Errorf("未知的工具调用: %s", toolCall.Function.Name)
	}
	return tool.ExecuteToolCall(ctx, robotCtx, toolCall)
}
