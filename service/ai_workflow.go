package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wechat-robot-client/interface/settings"
	"wechat-robot-client/model"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type ChatIntention string

const (
	ChatIntentionSing             ChatIntention = "sing"
	ChatIntentionSongRequest      ChatIntention = "song_request"
	ChatIntentionDrawAPicture     ChatIntention = "draw_a_picture"
	ChatIntentionEditPictures     ChatIntention = "edit_pictures"
	ChatIntentionImageRecognizer  ChatIntention = "image_recognizer"
	ChatIntentionTTS              ChatIntention = "tts"
	ChatIntentionLTTS             ChatIntention = "ltts"
	ChatIntentionDYVideoParse     ChatIntention = "dy_video_parse"
	ChatIntentionApplyToJoinGroup ChatIntention = "apply_to_join_group"
	ChatIntentionAIDisabled       ChatIntention = "ai_disabled"
	ChatIntentionChat             ChatIntention = "chat"
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

type TTSText struct {
	Text string `json:"text"`
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

func (s *AIWorkflowService) ChatIntentionSimple(message string, referMessage *model.Message) (bool, ChatIntention) {
	if strings.Contains(message, "https://v.douyin.com") {
		return true, ChatIntentionDYVideoParse
	}
	if message == "#关闭AI" {
		return true, ChatIntentionAIDisabled
	}
	return false, ChatIntentionChat
}

func (s *AIWorkflowService) ChatIntention(message string, referMessage *model.Message) ChatIntention {
	// 简单意图分析
	matched, intention := s.ChatIntentionSimple(message, referMessage)
	if matched {
		return intention
	}

	// 复杂意图分析
	aiConfig := s.config.GetAIConfig()
	openaiConfig := openai.DefaultConfig(aiConfig.APIKey)
	openaiConfig.BaseURL = aiConfig.BaseURL

	client := openai.NewClientWithConfig(openaiConfig)

	var commonSystemMessage = "你现在是一个专业的需求分析师，能够精准的识别用户的需求。请根据用户的输入内容，判断用户的意图，并返回意图分类结果。"
	var systemMessage string
	var enums []string

	if referMessage == nil {
		systemMessage = `意图分类结果包括以下几种：
1. sing：用户想要唱歌。
2. song_request：用户想要点歌。
3. draw_a_picture：用户想要画画。
4. edit_pictures：用户想要编辑图片。
5. image_recognizer：用户想要识别图片内容。
6. ltts：用户想要将长文本转换为语音。
7. tts：用户想要将文本转换为语音，注意区分ltts，如果用户没有明确说明，则默认为tts。
8. chat：用户想要闲聊。
如果无法识别意图，那就归类为闲聊：chat。`
		enums = []string{"sing", "song_request", "draw_a_picture", "edit_pictures", "image_recognizer", "ltts", "tts", "chat"}
	} else {
		switch referMessage.Type {
		case model.MsgTypeText:
			systemMessage = `意图分类结果包括以下几种：
1. sing：用户想要唱歌。
2. song_request：用户想要点歌。
3. draw_a_picture：用户想要画画。
4. tts：用户想要将文本转换为语音。
5. dy_video_parse: 用户想要解析抖音视频链接，下载抖音视频，解析抖音视频，解析抖音链接。
6. apply_to_join_group：申请进群，用户想要加入某个群聊。
7. chat：用户想要闲聊。
前面用户发来了一段文本，请根据当前用户的输入内容，判断用户想干什么，如果无法识别意图，那就归类为闲聊：chat。`
			enums = []string{"sing", "song_request", "draw_a_picture", "tts", "dy_video_parse", "chat", "apply_to_join_group"}
		case model.MsgTypeImage:
			systemMessage = `意图分类结果包括以下几种：
1. image_recognizer：用户想要识别图片内容。
2. edit_pictures：用户想要编辑图片。
前面用户发来了一张图片，请根据当前用户的输入内容，判断用户想干什么。`
			enums = []string{"image_recognizer", "edit_pictures"}
		case model.MsgTypeApp:
			if referMessage.AppMsgType == model.AppMsgTypeAttach {
				systemMessage = `意图分类结果包括以下几种：
1. ltts：用户想要将长文本转换为语音。
2. chat：用户想要闲聊。
前面用户发来了一个文件，请根据当前用户的输入内容，判断用户想干什么，如果无法识别意图，那就归类为闲聊：chat。`
				enums = []string{"ltts", "chat"}
			}
		}
	}

	if len(enums) == 0 {
		// 未支持的消息类型
		return ChatIntentionChat
	}

	aiMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("%s\n%s", commonSystemMessage, systemMessage),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: message,
		},
	}

	var result ChatCategories
	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"class_name": {
				Type:        jsonschema.String,
				Enum:        enums,
				Description: "用户意图分类",
			},
		},
		Required:             []string{"class_name"},
		AdditionalProperties: false,
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    aiConfig.WorkflowModel,
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

func (s *AIWorkflowService) GetSongRequestTitle(message string) string {
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
			Content: message,
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
			Model:    aiConfig.WorkflowModel,
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

func (s *AIWorkflowService) GetDrawingPrompt(message string) string {
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
			Content: message,
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
			Model:    aiConfig.WorkflowModel,
			Messages: aiMessages,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:        "drawing_prompt",
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

func (s *AIWorkflowService) GetTTSText(message string) string {
	aiConfig := s.config.GetAIConfig()
	openaiConfig := openai.DefaultConfig(aiConfig.APIKey)
	openaiConfig.BaseURL = aiConfig.BaseURL

	client := openai.NewClientWithConfig(openaiConfig)

	aiMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: `用户现在想文本转语音，根据用户输入的内容，提取用户想要将转换成语音的那部份内容。`,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: message,
		},
	}

	var result TTSText
	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"text": {
				Type:        jsonschema.String,
				Description: "用户想要转换的文本",
			},
		},
		Required:             []string{"text"},
		AdditionalProperties: false,
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    aiConfig.WorkflowModel,
			Messages: aiMessages,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:        "tts_text",
					Description: "用户现在想文本转语音，根据用户输入的内容，提取用户想要将转换成语音的那部份内容。",
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

	return result.Text
}
