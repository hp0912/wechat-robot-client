package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"wechat-robot-client/model"
	"wechat-robot-client/utils"
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

func newOpenAIClient(apiKey, baseURL string) openai.Client {
	return openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(utils.NormalizeAIBaseURL(baseURL)),
	)
}

func streamChatCompletionMessage(ctx context.Context, client *openai.Client, req openai.ChatCompletionNewParams) (openai.ChatCompletionMessage, error) {
	stream := client.Chat.Completions.NewStreaming(ctx, req)
	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		acc.AddChunk(stream.Current())
	}
	if err := stream.Err(); err != nil {
		return openai.ChatCompletionMessage{}, err
	}
	if len(acc.Choices) == 0 {
		return openai.ChatCompletionMessage{}, fmt.Errorf("empty response")
	}
	return acc.Choices[0].Message, nil
}

func (s *AIMomentService) GetMomentMood(content string, momentSettings model.MomentSettings) *MomentMood {
	client := newOpenAIClient(momentSettings.AIAPIKey, momentSettings.AIBaseURL)

	aiMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(`朋友在社交平台上发了一条动态，请根据动态内容，判断这条动态是否适合点赞和评论：
如果朋友发了一条开心的动态，那么适合点赞和评论，表示祝贺。
如果朋友发了一条悲伤的动态，那么适合评论，表示安慰，但是不适合点赞。
举个例子，如果朋友生病了，那么适合评论，表示安慰，但是不适合点赞。
如果好友发布的是亲人去世或者某个知名人士去世的消息，则统一按照不适合点赞和评论处理。		
`),
		openai.UserMessage(content),
	}

	var result MomentMood
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"like": map[string]any{
				"type":        "string",
				"enum":        []string{"yes", "no"},
				"description": "是否适合点赞，适合点赞返回 'yes'，不适合点赞返回 'no'",
			},
			"comment": map[string]any{
				"type":        "string",
				"enum":        []string{"yes", "no"},
				"description": "是否适合评论，适合评论返回 'yes'，不适合评论返回 'no'",
			},
		},
		"required":             []string{"like", "comment"},
		"additionalProperties": false,
	}

	msg, err := streamChatCompletionMessage(
		context.Background(),
		&client,
		openai.ChatCompletionNewParams{
			Model:    momentSettings.WorkflowModel,
			Messages: aiMessages,
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
					JSONSchema: openai.ResponseFormatJSONSchemaJSONSchemaParam{
						Name:        "moment_mood",
						Description: openai.String("根据好友发布的动态内容，判断这条动态是否适合点赞和评论。"),
						Strict:      openai.Bool(true),
						Schema:      schema,
					},
				},
			},
		},
	)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal([]byte(msg.Content), &result); err != nil {
		return nil
	}

	return &result
}

func (s *AIMomentService) Comment(content string, momentSettings model.MomentSettings) (openai.ChatCompletionMessage, error) {
	systemPrompt := momentSettings.CommentPrompt
	if momentSettings.MaxCompletionTokens != nil && *momentSettings.MaxCompletionTokens > 0 {
		systemPrompt += fmt.Sprintf("\n\n请注意，每次回答不能超过%d个汉字。", *momentSettings.MaxCompletionTokens)
	}

	aiMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(content),
	}

	client := newOpenAIClient(momentSettings.AIAPIKey, momentSettings.AIBaseURL)
	req := openai.ChatCompletionNewParams{
		Model:    momentSettings.CommentModel,
		Messages: aiMessages,
	}

	assistantMsg, err := streamChatCompletionMessage(context.Background(), &client, req)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}

	if assistantMsg.Content == "" {
		return openai.ChatCompletionMessage{}, fmt.Errorf("AI返回了空内容，请联系管理员")
	}
	return assistantMsg, nil
}
