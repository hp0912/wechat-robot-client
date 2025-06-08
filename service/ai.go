package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
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
	ctx              context.Context
	chatRoomSettings *model.ChatRoomSettings
	globalSettings   *model.GlobalSettings
	friendSettings   *model.FriendSettings
}

const defaultTTL = 10 * time.Minute

func NewAIService(ctx context.Context, message *model.Message) *AIService {
	gsRespo := repository.NewGlobalSettingsRepo(ctx, vars.DB)
	crsRespo := repository.NewChatRoomSettingsRepo(ctx, vars.DB)
	fsRespo := repository.NewFriendSettingsRepo(ctx, vars.DB)
	var globalSettings *model.GlobalSettings
	var chatRoomSettings *model.ChatRoomSettings
	var friendSettings *model.FriendSettings
	globalSettings, _ = gsRespo.GetGlobalSettings()
	if message.IsChatRoom {
		chatRoomSettings, _ = crsRespo.GetChatRoomSettings(message.FromWxID)
	} else {
		friendSettings, _ = fsRespo.GetFriendSettings(message.FromWxID)
	}
	return &AIService{
		ctx:              ctx,
		globalSettings:   globalSettings,
		chatRoomSettings: chatRoomSettings,
		friendSettings:   friendSettings,
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

func (s *AIService) GetAIConfig() (baseURL string, apiKey string, model string, prompt string) {
	if s.globalSettings != nil {
		if s.globalSettings.ChatBaseURL != "" {
			baseURL = s.globalSettings.ChatBaseURL
		}
		if s.globalSettings.ChatAPIKey != "" {
			apiKey = s.globalSettings.ChatAPIKey
		}
		if s.globalSettings.ChatModel != "" {
			model = s.globalSettings.ChatModel
		}
		if s.globalSettings.ChatPrompt != "" {
			prompt = s.globalSettings.ChatPrompt
		}
	}
	if s.chatRoomSettings != nil {
		if s.chatRoomSettings.ChatBaseURL != nil && *s.chatRoomSettings.ChatBaseURL != "" {
			baseURL = *s.chatRoomSettings.ChatBaseURL
		}
		if s.chatRoomSettings.ChatAPIKey != nil && *s.chatRoomSettings.ChatAPIKey != "" {
			apiKey = *s.chatRoomSettings.ChatAPIKey
		}
		if s.chatRoomSettings.ChatModel != nil && *s.chatRoomSettings.ChatModel != "" {
			model = *s.chatRoomSettings.ChatModel
		}
		if s.chatRoomSettings.ChatPrompt != nil && *s.chatRoomSettings.ChatPrompt != "" {
			prompt = *s.chatRoomSettings.ChatPrompt
		}
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL += "/v1"
	}
	return
}

func (s *AIService) IsAIEnabled() bool {
	if s.friendSettings != nil && s.friendSettings.ChatAIEnabled != nil {
		return *s.friendSettings.ChatAIEnabled
	}
	if s.globalSettings != nil && s.globalSettings.ChatAIEnabled != nil {
		return *s.globalSettings.ChatAIEnabled
	}
	return false
}

func (s *AIService) IsAITrigger(message *model.Message) bool {
	if message.IsAtMe {
		re := regexp.MustCompile(vars.TrimAtRegexp)
		message.Content = re.ReplaceAllString(message.Content, "")
		return true
	}
	if s.chatRoomSettings == nil {
		if s.globalSettings == nil {
			return false
		}
		if s.globalSettings.ChatAIEnabled == nil || !*s.globalSettings.ChatAIEnabled {
			return false
		}
		isAITrigger := *s.globalSettings.ChatAITrigger != "" && strings.HasPrefix(message.Content, *s.globalSettings.ChatAITrigger)
		if isAITrigger {
			re := regexp.MustCompile(regexp.QuoteMeta(*s.globalSettings.ChatAITrigger) + `[\s，,：:]*`)
			message.Content = re.ReplaceAllString(message.Content, "")
		}
		return isAITrigger
	}
	if s.chatRoomSettings.ChatAIEnabled == nil || !*s.chatRoomSettings.ChatAIEnabled {
		return false
	}
	if s.chatRoomSettings.ChatAITrigger != nil && *s.chatRoomSettings.ChatAITrigger != "" {
		isAITrigger := *s.chatRoomSettings.ChatAITrigger != "" && strings.HasPrefix(message.Content, *s.chatRoomSettings.ChatAITrigger)
		if isAITrigger {
			re := regexp.MustCompile(regexp.QuoteMeta(*s.chatRoomSettings.ChatAITrigger) + `[\s，,：:]*`)
			message.Content = re.ReplaceAllString(message.Content, "")
		}
		return isAITrigger
	}
	isAITrigger := s.globalSettings != nil && s.globalSettings.ChatAITrigger != nil && *s.globalSettings.ChatAITrigger != "" &&
		strings.HasPrefix(message.Content, *s.globalSettings.ChatAITrigger)
	if isAITrigger {
		re := regexp.MustCompile(regexp.QuoteMeta(*s.globalSettings.ChatAITrigger) + `[\s，,：:]*`)
		message.Content = re.ReplaceAllString(message.Content, "")
	}
	return isAITrigger
}

func (s *AIService) ChatIntention(message *model.Message) ChatIntention {
	baseURL, apiKey, model, _ := s.GetAIConfig()
	aiConfig := openai.DefaultConfig(apiKey)
	aiConfig.BaseURL = baseURL

	client := openai.NewClientWithConfig(aiConfig)

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
			Model:    model,
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
	baseURL, apiKey, model, _ := s.GetAIConfig()
	aiConfig := openai.DefaultConfig(apiKey)
	aiConfig.BaseURL = baseURL

	client := openai.NewClientWithConfig(aiConfig)

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
			Model:    model,
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
	baseURL, apiKey, model, prompt := s.GetAIConfig()
	if prompt != "" {
		systemMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		}
		aiMessages = append([]openai.ChatCompletionMessage{systemMessage}, aiMessages...)
	}
	aiConfig := openai.DefaultConfig(apiKey)
	aiConfig.BaseURL = baseURL
	client := openai.NewClientWithConfig(aiConfig)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
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
