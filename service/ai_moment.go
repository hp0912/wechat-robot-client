package service

import (
	"context"
	"fmt"
	"wechat-robot-client/model"
	"wechat-robot-client/utils"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type AIMomentService struct {
	ctx context.Context
}

type MomentMood struct {
	Like    string `json:"like"`
	Comment string `json:"comment"`
}

func NewAIMomentService(ctx context.Context) *AIMomentService {
	return &AIMomentService{
		ctx: ctx,
	}
}

func (s *AIMomentService) GetMomentMood(content string, momentSettings model.MomentSettings) *MomentMood {
	openaiConfig := openai.DefaultConfig(momentSettings.AIAPIKey)
	openaiConfig.BaseURL = momentSettings.AIBaseURL
	openaiConfig.BaseURL = utils.NormalizeAIBaseURL(openaiConfig.BaseURL)

	client := openai.NewClientWithConfig(openaiConfig)

	aiMessages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `朋友在社交平台上发了一条动态，请根据动态内容，判断这条动态是否适合点赞和评论：
如果朋友发了一条开心的动态，那么适合点赞和评论，表示祝贺。
如果朋友发了一条悲伤的动态，那么适合评论，表示安慰，但是不适合点赞。
举个例子，如果朋友生病了，那么适合评论，表示安慰，但是不适合点赞。
如果好友发布的是亲人去世或者某个知名人士去世的消息，则统一按照不适合点赞和评论处理。		
`,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: content,
		},
	}

	var result MomentMood
	schema := &jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"like": {
				Type:        jsonschema.String,
				Enum:        []string{"yes", "no"},
				Description: "是否适合点赞，适合点赞返回 'yes'，不适合点赞返回 'no'",
			},
			"comment": {
				Type:        jsonschema.String,
				Enum:        []string{"yes", "no"},
				Description: "是否适合评论，适合评论返回 'yes'，不适合评论返回 'no'",
			},
		},
		Required:             []string{"like", "comment"},
		AdditionalProperties: false,
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    momentSettings.WorkflowModel,
			Messages: aiMessages,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:        "moment_mood",
					Description: "根据好友发布的动态内容，判断这条动态是否适合点赞和评论。",
					Strict:      true,
					Schema:      schema,
				},
			},
			Stream: false,
		},
	)
	if err != nil {
		return nil
	}
	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		return nil
	}

	return &result
}

func (s *AIMomentService) Comment(content string, momentSettings model.MomentSettings) (openai.ChatCompletionMessage, error) {
	systemPrompt := momentSettings.CommentPrompt
	if momentSettings.MaxCompletionTokens != nil && *momentSettings.MaxCompletionTokens > 0 {
		systemPrompt += fmt.Sprintf("\n\n请注意，每次回答不能超过%d个汉字。", *momentSettings.MaxCompletionTokens)
	}

	aiMessages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: content,
		},
	}

	openaiConfig := openai.DefaultConfig(momentSettings.AIAPIKey)
	openaiConfig.BaseURL = momentSettings.AIBaseURL
	openaiConfig.BaseURL = utils.NormalizeAIBaseURL(openaiConfig.BaseURL)

	client := openai.NewClientWithConfig(openaiConfig)
	req := openai.ChatCompletionRequest{
		Model:    momentSettings.CommentModel,
		Messages: aiMessages,
		Stream:   false,
	}
	if momentSettings.MaxCompletionTokens != nil && *momentSettings.MaxCompletionTokens > 0 {
		req.MaxCompletionTokens = *momentSettings.MaxCompletionTokens
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
