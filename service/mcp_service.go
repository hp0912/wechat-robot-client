package service

import (
	"context"
	"fmt"
	"log"

	"wechat-robot-client/interface/ai"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/mcp"
	"wechat-robot-client/repository"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

type MCPService struct {
	ctx           context.Context
	db            *gorm.DB
	manager       *mcp.MCPManager
	converter     *mcp.MCPToolConverter
	mcpServerRepo *repository.MCPServer
}

var _ ai.MCPService = (*MCPService)(nil)

func NewMCPService(ctx context.Context, db *gorm.DB, messageSender mcp.MessageSender) *MCPService {
	manager := mcp.NewMCPManager(db)
	converter := mcp.NewMCPToolConverter(manager, messageSender)
	repo := repository.NewMCPServerRepo(ctx, db)

	return &MCPService{
		ctx:           ctx,
		db:            db,
		manager:       manager,
		converter:     converter,
		mcpServerRepo: repo,
	}
}

func (s *MCPService) Name() string {
	return "MCPService"
}

func (s *MCPService) Initialize() error {
	log.Println("Initializing MCP Service...")
	return s.manager.Initialize()
}

func (s *MCPService) Shutdown(ctx context.Context) error {
	return s.manager.Shutdown()
}

// GetAllTools 获取所有可用工具（OpenAI格式）
func (s *MCPService) GetAllTools() ([]openai.Tool, error) {
	return s.converter.ConvertMCPToolsToOpenAI(s.ctx)
}

// GetToolsByServerName 获取指定服务器的工具
func (s *MCPService) GetToolsByServerName(serverName string) ([]openai.Tool, error) {
	return s.converter.GetToolsByServer(s.ctx, serverName)
}

// GetToolsByServerID 获取指定MCP服务器提供的工具列表（MCP SDK原始格式）
func (s *MCPService) GetToolsByServerID(serverID uint64) ([]*sdkmcp.Tool, error) {
	// 检查服务器是否存在且已启用
	server, err := s.mcpServerRepo.FindByID(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to find server: %w", err)
	}
	if server == nil {
		return nil, fmt.Errorf("server not found")
	}
	if server.Enabled == nil || !*server.Enabled {
		return nil, fmt.Errorf("server is not enabled")
	}

	// 通过manager获取客户端并列出工具
	client, err := s.manager.GetClient(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	tools, err := client.ListTools(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return tools, nil
}

// ExecuteToolCall 执行工具调用
func (s *MCPService) ExecuteToolCall(robotCtx mcp.RobotContext, toolCall openai.ToolCall) (string, bool, error) {
	return s.converter.ExecuteOpenAIToolCall(s.ctx, robotCtx, toolCall)
}

// ChatWithMCPTools AI聊天（带MCP工具支持）
func (s *MCPService) ChatWithMCPTools(
	robotCtx mcp.RobotContext,
	client *openai.Client,
	req openai.ChatCompletionRequest,
	maxIterations int,
) (openai.ChatCompletionMessage, error) {
	if maxIterations <= 0 {
		maxIterations = 5 // 默认最多5轮工具调用
	}

	// 获取所有可用工具
	tools, err := s.GetAllTools()
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("failed to get tools: %w", err)
	}

	// 如果没有可用工具，直接调用AI
	if len(tools) == 0 {
		return s.chatWithoutTools(client, req)
	}

	req.Tools = tools

	// 构建包含工具描述的系统提示词
	if len(req.Messages) > 0 && req.Messages[0].Role == openai.ChatMessageRoleSystem {
		enhancedPrompt, err := s.converter.BuildSystemPromptWithMCPTools(s.ctx, req.Messages[0].Content)
		if err == nil {
			req.Messages[0].Content = enhancedPrompt
		}
	}

	// 迭代调用，支持多轮工具调用
	for i := 0; i < maxIterations; i++ {
		// 调用AI
		resp, err := client.CreateChatCompletion(s.ctx, req)
		if err != nil {
			return openai.ChatCompletionMessage{}, fmt.Errorf("failed to call ai: %w", err)
		}

		if len(resp.Choices) == 0 {
			return openai.ChatCompletionMessage{}, fmt.Errorf("ai returned empty response")
		}

		assistantMsg := resp.Choices[0].Message

		// 检查是否需要调用工具
		if len(assistantMsg.ToolCalls) == 0 {
			// 没有工具调用，返回结果
			return assistantMsg, nil
		}

		// 将assistant消息添加到历史
		req.Messages = append(req.Messages, assistantMsg)

		// 执行所有工具调用
		for _, toolCall := range assistantMsg.ToolCalls {
			log.Printf("Executing tool: %s", toolCall.Function.Name)

			// 执行工具
			result, immediately, err := s.ExecuteToolCall(robotCtx, toolCall)
			if err == nil {
				// 工具调用结果立即返回
				if immediately {
					return openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleAssistant,
						Content:    result,
						ToolCallID: toolCall.ID,
					}, nil
				}
			} else {
				// 工具执行失败，返回错误信息
				result = err.Error()
				log.Println(result)
			}

			// 将工具结果添加到消息历史
			toolResultMsg := openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			}
			req.Messages = append(req.Messages, toolResultMsg)
		}
	}

	// 达到最大迭代次数，返回最后的消息
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		if lastMsg.Role == openai.ChatMessageRoleAssistant {
			return lastMsg, nil
		}
	}

	return openai.ChatCompletionMessage{}, fmt.Errorf("max iterations reached without final answer")
}

// chatWithoutTools 不使用工具的简单聊天
func (s *MCPService) chatWithoutTools(
	client *openai.Client,
	req openai.ChatCompletionRequest,
) (openai.ChatCompletionMessage, error) {
	resp, err := client.CreateChatCompletion(s.ctx, req)
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}

	if len(resp.Choices) == 0 {
		return openai.ChatCompletionMessage{}, fmt.Errorf("ai returned empty response")
	}

	return resp.Choices[0].Message, nil
}

// AddServer 添加MCP服务器
func (s *MCPService) AddServer(server *model.MCPServer) error {
	// 保存到数据库
	if err := s.mcpServerRepo.Create(server); err != nil {
		return fmt.Errorf("failed to create server in database: %w", err)
	}

	// 如果启用，则连接
	if server.Enabled != nil && *server.Enabled {
		if err := s.manager.AddServer(server); err != nil {
			return fmt.Errorf("failed to connect server: %w", err)
		}
	}

	return nil
}

// RemoveServer 移除MCP服务器
func (s *MCPService) RemoveServer(serverID uint64) error {
	// 从管理器移除
	s.manager.RemoveServer(serverID)

	// 从数据库删除
	return s.mcpServerRepo.Delete(serverID)
}

// UpdateServer 更新MCP服务器配置
func (s *MCPService) UpdateServer(server *model.MCPServer) error {
	// 更新数据库
	if err := s.mcpServerRepo.Update(server); err != nil {
		return fmt.Errorf("failed to update server in database: %w", err)
	}

	// 重新加载
	if server.Enabled != nil && *server.Enabled {
		if err := s.manager.ReloadServer(server.ID); err != nil {
			return fmt.Errorf("failed to reload server: %w", err)
		}
	} else {
		// 如果禁用，则断开连接
		s.manager.RemoveServer(server.ID)
	}

	return nil
}

// EnableServer 启用服务器
func (s *MCPService) EnableServer(serverID uint64) error {
	return s.manager.EnableServer(serverID)
}

// DisableServer 禁用服务器
func (s *MCPService) DisableServer(serverID uint64) error {
	return s.manager.DisableServer(serverID)
}

// GetServerByID 获取服务器配置
func (s *MCPService) GetServerByID(serverID uint64) (*model.MCPServer, error) {
	return s.mcpServerRepo.FindByID(serverID)
}

// GetAllServers 获取所有服务器配置
func (s *MCPService) GetAllServers() ([]*model.MCPServer, error) {
	return s.mcpServerRepo.FindAll()
}

// GetEnabledServers 获取已启用的服务器
func (s *MCPService) GetEnabledServers() ([]*model.MCPServer, error) {
	return s.mcpServerRepo.FindEnabled()
}

// GetServerStats 获取服务器统计信息
func (s *MCPService) GetServerStats(serverID uint64) (*mcp.MCPConnectionStats, error) {
	return s.manager.GetServerStats(serverID)
}

// GetActiveServerCount 获取活动服务器数量
func (s *MCPService) GetActiveServerCount() int {
	return s.manager.GetActiveServerCount()
}

// ReloadServer 重新加载服务器
func (s *MCPService) ReloadServer(serverID uint64) error {
	return s.manager.ReloadServer(serverID)
}

// TestServerConnection 测试服务器连接
func (s *MCPService) TestServerConnection(server *model.MCPServer) error {
	// 创建临时客户端
	client, err := mcp.NewMCPClient(server)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// 尝试连接
	ctx, cancel := context.WithTimeout(s.ctx, 30)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Disconnect()

	// 尝试初始化
	if _, err := client.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	return nil
}
