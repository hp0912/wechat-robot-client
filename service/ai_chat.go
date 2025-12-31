package service

import (
	"context"
	"fmt"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/pkg/mcp"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
)

type AIChatService struct {
	ctx    context.Context
	config settings.Settings
}

func NewAIChatService(ctx context.Context, config settings.Settings) *AIChatService {
	return &AIChatService{
		ctx:    ctx,
		config: config,
	}
}

func (s *AIChatService) Chat(robotCtx mcp.RobotContext, aiMessages []openai.ChatCompletionMessage) (openai.ChatCompletionMessage, error) {
	aiConfig := s.config.GetAIConfig()
	if aiConfig.Prompt != "" {
		systemMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: aiConfig.Prompt,
		}
		if aiConfig.MaxCompletionTokens > 0 {
			systemMessage.Content += fmt.Sprintf("\n\n请注意，每次回答不能超过%d个汉字。", aiConfig.MaxCompletionTokens)
		}
		aiMessages = append([]openai.ChatCompletionMessage{systemMessage}, aiMessages...)
	}
	openaiConfig := openai.DefaultConfig(aiConfig.APIKey)
	openaiConfig.BaseURL = aiConfig.BaseURL
	client := openai.NewClientWithConfig(openaiConfig)
	req := openai.ChatCompletionRequest{
		Model:    aiConfig.Model,
		Messages: aiMessages,
		Stream:   false,
	}
	if aiConfig.MaxCompletionTokens > 0 {
		req.MaxCompletionTokens = aiConfig.MaxCompletionTokens
	}
	return vars.MCPService.ChatWithMCPTools(robotCtx, client, req, 0)
}
