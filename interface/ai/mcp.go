package ai

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sashabaranov/go-openai"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/mcp"
)

type MCPService interface {
	Name() string
	Initialize() error
	Shutdown(ctx context.Context) error
	GetAllTools() ([]openai.Tool, error)
	GetToolsByServerName(serverName string) ([]openai.Tool, error)
	GetToolsByServerID(serverID uint64) ([]*sdkmcp.Tool, error)
	ExecuteToolCall(robotCtx mcp.RobotContext, toolCall openai.ToolCall) (string, bool, error)
	ChatWithMCPTools(
		robotCtx mcp.RobotContext,
		client *openai.Client,
		req openai.ChatCompletionRequest,
		maxIterations int,
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
