package ai

import (
	"context"

	"github.com/openai/openai-go/v3"

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
	GetAllTools(robotCtx *robotctx.RobotContext) ([]openai.ChatCompletionToolUnionParam, error)
	ChatWithTools(
		robotCtx *robotctx.RobotContext,
		client *openai.Client,
		req openai.ChatCompletionNewParams,
	) (openai.ChatCompletionMessage, error)
}
