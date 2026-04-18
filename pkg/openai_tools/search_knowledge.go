package openaitools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"

	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type SearchKnowledgeTool struct {
	KnowledgeService   ai.KnowledgeService
	cachedSystemPrompt string
}

func NewSearchKnowledgeTool(knowledgeService ai.KnowledgeService) OpenAITool {
	return &SearchKnowledgeTool{
		KnowledgeService: knowledgeService,
	}
}

func (t *SearchKnowledgeTool) GetOpenAITool(robotCtx *robotctx.RobotContext) *openai.Tool {
	systemPrompt, err := t.BuildSystemPrompt(context.Background(), robotCtx)
	if err != nil {
		fmt.Printf("构建系统提示词失败: %v\n", err)
		return nil
	}
	if systemPrompt == "" {
		return nil
	}
	return &openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "search_document",
			Description: "检索当前群聊绑定的文档，根据用户问题语义搜索最相关的知识内容。只有当用户的问题可能与群聊绑定的文档主题相关时，才调用此工具获取准确信息。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]string{
						"type":        "string",
						"description": "用于检索文档的查询语句，应当是用户问题的核心关键词或语义描述",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

func (t *SearchKnowledgeTool) BuildSystemPrompt(ctx context.Context, robotCtx *robotctx.RobotContext) (string, error) {
	if t.cachedSystemPrompt != "" {
		systemPrompt := t.cachedSystemPrompt
		t.cachedSystemPrompt = "" // 使用一次后清空缓存，确保后续能获取到最新的群聊设置和知识库绑定信息
		return systemPrompt, nil
	}

	if t.KnowledgeService == nil || !strings.HasSuffix(robotCtx.FromWxID, "@chatroom") {
		return "", nil
	}

	crsRepo := repository.NewChatRoomSettingsRepo(ctx, vars.DB)
	chatRoomSettings, err := crsRepo.GetChatRoomSettings(robotCtx.FromWxID)
	if err != nil {
		return "", fmt.Errorf("获取群聊设置失败: %w", err)
	}
	if chatRoomSettings == nil {
		return "", nil
	}

	codes, err := chatRoomSettings.GetKnowledgeCategoryCodes()
	if err != nil {
		return "", fmt.Errorf("解析知识库分类失败: %w", err)
	}
	if len(codes) == 0 {
		return "", nil
	}

	normalized := make([]string, 0, len(codes))
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		normalized = append(normalized, code)
	}
	if len(normalized) == 0 {
		return "", nil
	}

	categoryRepo := repository.NewKnowledgeCategoryRepo(ctx, vars.DB)
	categories, err := categoryRepo.GetByCodes(codes)
	if err != nil {
		return "", fmt.Errorf("获取知识库分类信息失败: %w", err)
	}
	if len(categories) == 0 {
		return "", nil
	}

	categoryByCode := make(map[string]*model.KnowledgeCategory, len(categories))
	for _, category := range categories {
		categoryByCode[category.Code] = category
	}

	var sb strings.Builder
	sb.WriteString("\n\n## 下面是群聊可用的文档内容概览:\n")
	sb.WriteString("**只有当用户查询的信息在文档的覆盖范围内时，才调用 `search_document` 工具来检索文档获取准确信息，而不是凭记忆回答，并且你不要暴露你的知识是从文档获取的。**\n\n")
	validCodes := make([]string, 0, len(categories))
	for _, code := range codes {
		category, ok := categoryByCode[code]
		if !ok || category.Type != model.KnowledgeCategoryTypeText {
			continue
		}
		desc := category.Description
		if desc == "" {
			desc = category.Name
		}
		fmt.Fprintf(&sb, "- **%s**（编码: %s）: %s\n", category.Name, category.Code, desc)
		validCodes = append(validCodes, category.Code)
	}

	if len(validCodes) == 0 {
		return "", nil
	}

	robotCtx.KnowledgeBaseCodes = validCodes
	t.cachedSystemPrompt = sb.String()

	return t.cachedSystemPrompt, nil
}

func (t *SearchKnowledgeTool) ExecuteToolCall(ctx context.Context, robotCtx *robotctx.RobotContext, toolCall openai.ToolCall) (string, bool, error) {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", false, fmt.Errorf("解析参数失败: %w", err)
	}
	if args.Query == "" {
		return "查询内容不能为空", false, nil
	}
	if t.KnowledgeService == nil {
		return "知识服务不可用", false, nil
	}
	if len(robotCtx.KnowledgeBaseCodes) == 0 {
		return "当前群聊未绑定任何知识库", false, nil
	}
	// 使用独立 context 避免捕获外层可能已过期的 ctx
	toolCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	results, err := t.KnowledgeService.SearchKnowledgeByCategories(toolCtx, args.Query, robotCtx.KnowledgeBaseCodes, 5)
	if err != nil {
		return fmt.Sprintf("检索知识库失败: %v", err), false, nil
	}
	if len(results) == 0 {
		return "未找到相关知识内容", false, nil
	}
	var sb strings.Builder
	sb.WriteString("以下是你获取到的知识，不要暴露你的知识是从文档获取的，而应该装作你自己知道的:\n\n")
	for i, doc := range results {
		title := doc.Payload["title"]
		content := doc.Payload["content"]
		category := doc.Payload["category"]
		if content != "" {
			fmt.Fprintf(&sb, "### %d. %s", i+1, title)
			if category != "" {
				fmt.Fprintf(&sb, "（分类: %s）", category)
			}
			sb.WriteString("\n")
			sb.WriteString(content)
			sb.WriteString("\n\n")
		}
	}
	return sb.String(), false, nil
}
