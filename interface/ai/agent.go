package ai

import (
	"context"

	"github.com/sashabaranov/go-openai"

	"wechat-robot-client/pkg/mcp"
	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/pkg/skills"
)

type AgentService interface {
	Name() string
	Initialize() error
	Shutdown(ctx context.Context) error
	GetMCPManager() *mcp.MCPManager
	GetSkillsManager() *skills.SkillsManager
	GetAllTools(robotCtx *robotctx.RobotContext) ([]openai.Tool, error)
	ChatWithTools(
		robotCtx robotctx.RobotContext,
		client *openai.Client,
		req openai.ChatCompletionRequest,
	) (openai.ChatCompletionMessage, error)
}
