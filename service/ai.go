package service

import (
	"context"
	"fmt"
	"time"
	"wechat-robot-client/model"
	"wechat-robot-client/vars"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type ChatIntention string

const (
	ChatIntentionSing         ChatIntention = "sing"
	ChatIntentionSongRequest  ChatIntention = "song_request"
	ChatIntentionDrawAPicture ChatIntention = "draw_a_picture"
	ChatIntentionEditPictures ChatIntention = "edit_pictures"
	ChatIntentionChat         ChatIntention = "chat"
)

type ChatCategories struct {
	ClassName ChatIntention `json:"class_name"`
}

type SongRequestMetadata struct {
	SongTitle string `json:"song_title"`
}

type AIService struct {
	ctx    context.Context
	config Settings
}

const defaultTTL = 10 * time.Minute

func NewAIService(ctx context.Context, config Settings) *AIService {
	return &AIService{
		ctx:    ctx,
		config: config,
	}
}

func (s *AIService) SetAISession(message *model.Message) error {
	return vars.RedisClient.Set(s.ctx, s.GetSessionID(message), true, defaultTTL).Err()
}

func (s *AIService) RenewAISession(message *model.Message) error {
	return vars.RedisClient.Expire(s.ctx, s.GetSessionID(message), defaultTTL).Err()
}

func (s *AIService) ExpireAISession(message *model.Message) error {
	return vars.RedisClient.Del(s.ctx, s.GetSessionID(message)).Err()
}

func (s *AIService) ExpireAllAISessionByChatRoomID(chatRoomID string) error {
	sessionID := fmt.Sprintf("ai_session_%s:", chatRoomID)
	keys, err := vars.RedisClient.Keys(s.ctx, sessionID+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return vars.RedisClient.Del(s.ctx, keys...).Err()
}

func (s *AIService) IsInAISession(message *model.Message) (bool, error) {
	cnt, err := vars.RedisClient.Exists(s.ctx, s.GetSessionID(message)).Result()
	return cnt == 1, err
}

func (s *AIService) GetSessionID(message *model.Message) string {
	return fmt.Sprintf("ai_session_%s:%s", message.FromWxID, message.SenderWxID)
}

func (s *AIService) IsAISessionStart(message *model.Message) bool {
	if message.Content == "#进入AI会话" {
		err := s.SetAISession(message)
		return err == nil
	}
	return false
}

func (s *AIService) IsAISessionEnd(message *model.Message) bool {
	if message.Content == "#退出AI会话" {
		err := s.ExpireAISession(message)
		return err == nil
	}
	return false
}

func (s *AIService) ChatIntention(message *model.Message) ChatIntention {
	aiConfig := s.config.GetAIConfig()
	openaiConfig := openai.DefaultConfig(aiConfig.APIKey)
	openaiConfig.BaseURL = aiConfig.BaseURL

	client := openai.NewClientWithConfig(openaiConfig)

	aiMessages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `你现在是一个专业的需求分析师，能够精准的识别用户的需求。
请根据用户的输入内容，判断用户的意图，并返回意图分类结果。
意图分类结果包括以下几种：
1. sing：用户想要唱歌。
2. song_request：用户想要点歌。
3. draw_a_picture：用户想要画画。
4. edit_pictures：用户想要编辑图片。
5. chat：用户想要闲聊。
如果无法识别意图，那就归类为闲聊：chat。`,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: message.Content,
		},
	}

	var result ChatCategories
	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"class_name": {
				Type:        jsonschema.String,
				Enum:        []string{"sing", "song_request", "draw_a_picture", "edit_pictures", "chat"},
				Description: "用户意图分类",
			},
		},
		Required:             []string{"class_name"},
		AdditionalProperties: false,
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    "gpt-4o-mini", // 固定写死，使用更小的模型以提高响应速度和降低成本
			Messages: aiMessages,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:        "chat_intention",
					Description: "你现在是一个专业的需求分析师，能够精准的识别用户的需求。",
					Strict:      true,
					Schema:      schema,
				},
			},
			Stream: false,
		},
	)
	if err != nil {
		return ChatIntentionChat
	}
	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		return ChatIntentionChat
	}

	return result.ClassName
}

func (s *AIService) GetSongRequestTitle(message *model.Message) string {
	aiConfig := s.config.GetAIConfig()
	openaiConfig := openai.DefaultConfig(aiConfig.APIKey)
	openaiConfig.BaseURL = aiConfig.BaseURL

	client := openai.NewClientWithConfig(openaiConfig)

	aiMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: `用户现在想点歌，请根据用户的输入内容，判断用户想要点的歌曲的歌名。`,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: message.Content,
		},
	}

	var result SongRequestMetadata
	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"song_title": {
				Type:        jsonschema.String,
				Description: "用户想要点的歌曲的歌名",
			},
		},
		Required:             []string{"song_title"},
		AdditionalProperties: false,
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    "gpt-4o-mini", // 固定写死，使用更小的模型以提高响应速度和降低成本
			Messages: aiMessages,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:        "song_request_title",
					Description: "用户现在想点歌，请根据用户的输入内容，判断用户想要点的歌曲的歌名。",
					Strict:      true,
					Schema:      schema,
				},
			},
			Stream: false,
		},
	)
	if err != nil {
		return ""
	}
	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		return ""
	}

	return result.SongTitle
}

func (s *AIService) Chat(aiMessages []openai.ChatCompletionMessage) (string, error) {
	aiConfig := s.config.GetAIConfig()
	if aiConfig.Prompt != "" {
		systemMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: aiConfig.Prompt,
		}
		aiMessages = append([]openai.ChatCompletionMessage{systemMessage}, aiMessages...)
	}
	openaiConfig := openai.DefaultConfig(aiConfig.APIKey)
	openaiConfig.BaseURL = aiConfig.BaseURL
	client := openai.NewClientWithConfig(openaiConfig)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    aiConfig.Model,
			Messages: aiMessages,
			Stream:   false,
		},
	)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("AI返回了空内容，请联系管理员")
	}
	return resp.Choices[0].Message.Content, nil
}
