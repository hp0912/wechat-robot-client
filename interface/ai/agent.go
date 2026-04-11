package ai

import (
	"context"

	"github.com/sashabaranov/go-openai"

	"wechat-robot-client/pkg/robotctx"
)

type AgentService interface {
	Name() string
	Initialize() error
	Shutdown(ctx context.Context) error
	GetAllTools() ([]openai.Tool, error)
	ChatWithTools(
		robotCtx robotctx.RobotContext,
		client *openai.Client,
		req openai.ChatCompletionRequest,
	) (openai.ChatCompletionMessage, error)
}
