package openaitools

import (
	"context"
	"wechat-robot-client/pkg/robotctx"

	"github.com/sashabaranov/go-openai"
)

type SearchKnowledgeTool struct {
}

func NewSearchKnowledgeTool() OpenAITool {
	return &SearchKnowledgeTool{}
}

func (t *SearchKnowledgeTool) GetOpenAITool() openai.Tool {
	return openai.Tool{}
}

func (t *SearchKnowledgeTool) BuildSystemPrompt(ctx context.Context, robotCtx robotctx.RobotContext) (string, error) {
	return "", nil
}

func (t *SearchKnowledgeTool) ExecuteToolCall(ctx context.Context, robotCtx robotctx.RobotContext, toolCall openai.ToolCall) (string, bool, error) {
	return "", false, nil
}
