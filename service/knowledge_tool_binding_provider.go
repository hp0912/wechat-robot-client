package service

import (
	"context"
	"fmt"
	"strings"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type KnowledgeToolBindingProvider struct{}

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
