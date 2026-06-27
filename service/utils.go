package service

import (
	"context"
	"fmt"
	"wechat-robot-client/utils"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

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
