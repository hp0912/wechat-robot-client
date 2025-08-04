package service

import (
	"context"
	"fmt"
	"wechat-robot-client/interface/ai"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
)

type AIChatService struct {
	ctx    context.Context
	config settings.Settings
}

var _ ai.AIService = (*AIChatService)(nil)

func NewAIChatService(ctx context.Context, config settings.Settings) *AIChatService {
	return &AIChatService{
		ctx:    ctx,
		config: config,
	}
}

func (s *AIChatService) SetAISession(message *model.Message) error {
	return vars.RedisClient.Set(s.ctx, s.GetSessionID(message), true, defaultTTL).Err()
}

func (s *AIChatService) RenewAISession(message *model.Message) error {
	return vars.RedisClient.Expire(s.ctx, s.GetSessionID(message), defaultTTL).Err()
}

func (s *AIChatService) ExpireAISession(message *model.Message) error {
	return vars.RedisClient.Del(s.ctx, s.GetSessionID(message)).Err()
}

func (s *AIChatService) ExpireAllAISessionByChatRoomID(chatRoomID string) error {
	sessionID := fmt.Sprintf("ai_chat_session_%s:", chatRoomID)
	keys, err := vars.RedisClient.Keys(s.ctx, sessionID+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return vars.RedisClient.Del(s.ctx, keys...).Err()
}

func (s *AIChatService) IsInAISession(message *model.Message) (bool, error) {
	cnt, err := vars.RedisClient.Exists(s.ctx, s.GetSessionID(message)).Result()
	return cnt == 1, err
}

func (s *AIChatService) GetSessionID(message *model.Message) string {
	return fmt.Sprintf("ai_chat_session_%s:%s", message.FromWxID, message.SenderWxID)
}

func (s *AIChatService) IsAISessionStart(message *model.Message) bool {
	if message.Content == "#进入AI会话" {
		err := s.SetAISession(message)
		return err == nil
	}
	return false
}

func (s *AIChatService) GetAISessionStartTips() string {
	return "AI会话已开始，请输入您的问题。10分钟不说话会话将自动结束，您也可以输入 #退出AI会话 来结束会话。"
}

func (s *AIChatService) IsAISessionEnd(message *model.Message) bool {
	if message.Content == "#退出AI会话" {
		err := s.ExpireAISession(message)
		return err == nil
	}
	return false
}

func (s *AIChatService) GetAISessionEndTips() string {
	return "AI会话已结束，您可以输入 #进入AI会话 来重新开始。"
}

func (s *AIChatService) Chat(aiMessages []openai.ChatCompletionMessage) (openai.ChatCompletionMessage, error) {
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
	// if hasImage {
	// 	req.Model = aiConfig.ImageRecognitionModel
	// }
	if aiConfig.MaxCompletionTokens > 0 {
		req.MaxCompletionTokens = aiConfig.MaxCompletionTokens
	}
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}
	if len(resp.Choices) == 0 {
		return openai.ChatCompletionMessage{}, fmt.Errorf("AI返回了空内容，请联系管理员")
	}
	return resp.Choices[0].Message, nil
}
