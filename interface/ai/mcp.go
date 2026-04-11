package ai

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sashabaranov/go-openai"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/mcp"
	"wechat-robot-client/pkg/robotctx"
)

type MCPService interface {
	Name() string
	Initialize() error
	Shutdown(ctx context.Context) error
	GetAllTools() ([]openai.Tool, error)
	GetToolsByServerName(serverName string) ([]openai.Tool, error)
	GetToolsByServerID(serverID uint64) ([]*sdkmcp.Tool, error)
	ExecuteToolCall(robotCtx robotctx.RobotContext, toolCall openai.ToolCall) (string, bool, error)
	ChatWithMCPTools(
		robotCtx robotctx.RobotContext,
		client *openai.Client,
		req openai.ChatCompletionRequest,
		maxIterations int,
		extraTools ...ExtraTool,
	) (openai.ChatCompletionMessage, error)
	AddServer(server *model.MCPServer) error
	RemoveServer(serverID uint64) error
	UpdateServer(server *model.MCPServer) error
	EnableServer(serverID uint64) error
	DisableServer(serverID uint64) error
	GetServerByID(serverID uint64) (*model.MCPServer, error)
	GetAllServers() ([]*model.MCPServer, error)
	GetEnabledServers() ([]*model.MCPServer, error)
	GetServerStats(serverID uint64) (*mcp.MCPConnectionStats, error)
	GetActiveServerCount() int
	ReloadServer(serverID uint64) error
	TestServerConnection(server *model.MCPServer) error
}

// ExtraTool 额外的内置工具（由调用方注入到 ChatWithMCPTools 中）
type ExtraTool struct {
	Tool    openai.Tool
	Handler func(toolCall openai.ToolCall) (result string, immediately bool, err error)
}
