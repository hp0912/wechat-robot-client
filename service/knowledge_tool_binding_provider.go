package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
)

type KnowledgeToolBindingProvider struct{}

var _ ai.ToolBindingProvider = (*KnowledgeToolBindingProvider)(nil)

func NewKnowledgeToolBindingProvider() *KnowledgeToolBindingProvider {
	return &KnowledgeToolBindingProvider{}
}

func (p *KnowledgeToolBindingProvider) Name() string {
	return "knowledge"
}

func (p *KnowledgeToolBindingProvider) BuildToolBinding(ctx context.Context, toolCtx ai.ToolBindingContext) (ai.ToolBinding, error) {
	if toolCtx.ChatRoomID == "" || vars.KnowledgeService == nil {
		return ai.ToolBinding{}, nil
	}

	// 优先使用上层预查询的群聊设置，避免重复查库
	chatRoomSettings := toolCtx.ChatRoomSettings
	if chatRoomSettings == nil {
		crsRepo := repository.NewChatRoomSettingsRepo(ctx, vars.DB)
		var err error
		chatRoomSettings, err = crsRepo.GetChatRoomSettings(toolCtx.ChatRoomID)
		if err != nil || chatRoomSettings == nil {
			return ai.ToolBinding{}, err
		}
	}

	codes, err := chatRoomSettings.GetKnowledgeCategoryCodes()
	if err != nil {
		return ai.ToolBinding{}, fmt.Errorf("parse knowledge categories: %w", err)
	}
	codes = normalizeKnowledgeCategoryCodes(codes)
	if len(codes) == 0 {
		return ai.ToolBinding{}, nil
	}

	categoryRepo := repository.NewKnowledgeCategoryRepo(ctx, vars.DB)
	categories, err := categoryRepo.GetByCodes(codes)
	if err != nil || len(categories) == 0 {
		return ai.ToolBinding{}, err
	}

	categoryByCode := make(map[string]*model.KnowledgeCategory, len(categories))
	for _, category := range categories {
		categoryByCode[category.Code] = category
	}

	var sb strings.Builder
	sb.WriteString("\n\n## 当前群聊可用的知识库:\n")
	sb.WriteString("以下是当前群聊绑定的知识库，当用户查询的信息与这些知识库的主题相关时，请调用 `search_knowledge` 工具来检索知识库获取准确信息，而不是凭记忆回答。\n\n")
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
		return ai.ToolBinding{}, nil
	}

	return ai.ToolBinding{
		Prompt: sb.String(),
		Tools:  []ai.ExtraTool{p.buildKnowledgeSearchTool(validCodes)},
	}, nil
}

func (p *KnowledgeToolBindingProvider) buildKnowledgeSearchTool(categories []string) ai.ExtraTool {
	return ai.ExtraTool{
		Tool: openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "search_knowledge",
				Description: "检索当前群聊绑定的知识库，根据用户问题语义搜索最相关的知识内容。当用户的问题可能与群聊绑定的知识库主题相关时，请调用此工具获取准确信息。",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]string{
							"type":        "string",
							"description": "用于检索知识库的查询语句，应当是用户问题的核心关键词或语义描述",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		Handler: func(toolCall openai.ToolCall) (string, bool, error) {
			var args struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return "", false, fmt.Errorf("解析参数失败: %w", err)
			}
			if args.Query == "" {
				return "查询内容不能为空", false, nil
			}
			// 使用独立 context 避免捕获外层可能已过期的 ctx
			toolCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			results, err := vars.KnowledgeService.SearchKnowledgeByCategories(toolCtx, args.Query, categories, 5)
			if err != nil {
				return fmt.Sprintf("检索知识库失败: %v", err), false, nil
			}
			if len(results) == 0 {
				return "未找到相关知识内容", false, nil
			}
			var sb strings.Builder
			sb.WriteString("以下是从知识库中检索到的相关内容:\n\n")
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
		},
	}
}
