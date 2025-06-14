package service

import (
	"context"
	"time"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/model"

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

type DrawingPrompt struct {
	Prompt string `json:"prompt"`
}

type AIWorkflowService struct {
	ctx    context.Context
	config settings.Settings
}

const defaultTTL = 10 * time.Minute

func NewAIWorkflowService(ctx context.Context, config settings.Settings) *AIWorkflowService {
	return &AIWorkflowService{
		ctx:    ctx,
		config: config,
	}
}

func (s *AIWorkflowService) ChatIntention(message *model.Message) ChatIntention {
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

func (s *AIWorkflowService) GetSongRequestTitle(message *model.Message) string {
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

func (s *AIWorkflowService) GetDrawingPrompt(message *model.Message) string {
	aiConfig := s.config.GetAIConfig()
	openaiConfig := openai.DefaultConfig(aiConfig.APIKey)
	openaiConfig.BaseURL = aiConfig.BaseURL

	client := openai.NewClientWithConfig(openaiConfig)

	aiMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: `用户现在想画画，请根据用户的输入内容，提取用户画画的提示词。`,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: message.Content,
		},
	}

	var result DrawingPrompt
	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"prompt": {
				Type:        jsonschema.String,
				Description: "用户画画的提示词",
			},
		},
		Required:             []string{"prompt"},
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
					Name:        "prompt",
					Description: "用户现在想画画，请根据用户的输入内容，提取用户画画的提示词。",
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

	return result.Prompt
}
