package openaitools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"

	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/vars"
)

type SearchMemoryTool struct{}

func NewSearchMemoryTool() OpenAITool {
	return &SearchMemoryTool{}
}

func (t *SearchMemoryTool) GetOpenAITool(robotCtx *robotctx.RobotContext) *openai.ChatCompletionToolUnionParam {
	if vars.MemoryService == nil {
		return nil
	}
	tool := openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "search_memory",
		Description: openai.String("搜索与当前用户相关的长期记忆。当用户的问题涉及历史对话、个人偏好、过往讨论、背景信息等需要回忆的内容时调用此工具。对于简单寒暄（如「你好」「在吗」）、纯实时问题（如「今天天气怎么样」）或无需历史上下文即可回答的问题，不要调用此工具。"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]string{
					"type":        "string",
					"description": "用于检索记忆的查询语句，应当是用户问题的核心语义，例如用户问「我之前提到过喜欢什么电影」，query 应为「喜欢什么电影」",
				},
			},
			"required": []string{"query"},
		},
	})
	return &tool
}

func (t *SearchMemoryTool) BuildSystemPrompt(ctx context.Context, robotCtx *robotctx.RobotContext) (string, error) {
	if vars.MemoryService == nil {
		return "", nil
	}
	return "" +
		"\n\n## 长期记忆查询工具\n" +
		"你可以调用 search_memory 工具来查询与当前对话相关的长期记忆。\n" +
		"**何时调用**：用户问题涉及过往对话、个人偏好、历史事件、背景信息、曾经提过的事情等需要回忆的内容。\n" +
		"**何时不调用**：简单寒暄（如「你好」「在吗」「晚安」）、纯实时信息查询（天气、新闻、股价）、无需任何历史上下文即可回答的常识问题。\n" +
		"参数 query 应为用户问题的核心语义表述，用于向量检索匹配相关记忆。", nil
}

func (t *SearchMemoryTool) ExecuteToolCall(ctx context.Context, robotCtx *robotctx.RobotContext, toolCall openai.ChatCompletionMessageToolCallUnion) (string, bool, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("解析参数失败: %w", err)
	}
	args.Query = strings.TrimSpace(args.Query)
	if args.Query == "" {
		return "查询内容不能为空", false, nil
	}
	if vars.MemoryService == nil {
		return "记忆服务不可用", false, nil
	}

	isChatRoom := strings.Contains(robotCtx.FromWxID, "@chatroom")
	memoryCtx := vars.MemoryService.BuildPromptContext(ctx, args.Query, robotCtx.FromWxID, robotCtx.SenderWxID, isChatRoom)
	if memoryCtx == "" {
		return "未找到相关长期记忆", false, nil
	}

	return fmt.Sprintf("以下是与当前对话相关的长期记忆，请参考这些信息回答用户问题：\n\n%s", memoryCtx), false, nil
}
