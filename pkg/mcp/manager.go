package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robotctx"
	"wechat-robot-client/repository"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

// MCPManager MCP服务管理器
type MCPManager struct {
	db               *gorm.DB
	clients          map[uint64]MCPClient
	heartbeatCancels map[uint64]context.CancelFunc
	mu               sync.RWMutex
	ctx              context.Context
	cancelFunc       context.CancelFunc
	repo             *repository.MCPServer
}

// NewMCPManager 创建MCP管理器
func NewMCPManager(db *gorm.DB) *MCPManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &MCPManager{
		db:               db,
		clients:          make(map[uint64]MCPClient),
		heartbeatCancels: make(map[uint64]context.CancelFunc),
		ctx:              ctx,
		cancelFunc:       cancel,
		repo:             repository.NewMCPServerRepo(ctx, db),
	}
}

// Initialize 初始化管理器，从数据库加载所有已启用的MCP服务器
func (m *MCPManager) Initialize() error {
	// 获取所有已启用的MCP服务器配置
	servers, err := m.repo.FindEnabled()
	if err != nil {
		return fmt.Errorf("failed to load mcp servers: %w", err)
	}

	log.Printf("Loading %d enabled MCP servers...", len(servers))

	// 创建并连接每个MCP服务器
	for _, server := range servers {
		if err := m.AddServer(server); err != nil {
			log.Printf("Failed to add MCP server %s: %v", server.Name, err)
			// 记录错误但继续加载其他服务器
			m.repo.UpdateConnectionError(server.ID, err.Error())
			continue
		}
	}

	log.Printf("MCP Manager initialized with %d active servers", m.GetActiveServerCount())
	return nil
}

// BuildSystemPrompt 构建包含MCP工具描述的系统提示词
func (m *MCPManager) BuildSystemPrompt(ctx context.Context, basePrompt string) (string, error) {
	allTools, err := m.GetAllTools(ctx)
	if err != nil {
		return basePrompt, err
	}

	if len(allTools) == 0 {
		return basePrompt, nil
	}

	intro := `你运行在一个支持 MCP（Model Context Protocol）工具的聊天应用环境中。
当你自身能力不足或需要访问外部数据时，应主动调用这些工具来完成任务。

1. 何时使用工具
- 当需要访问或处理「外部数据」时（例如：查询、点歌、统计、搜索、总结群聊内容等）。
- 当用户要求执行你自身无法完成的动作（例如：根据文本生成图片/音频/视频、图片内容提取或识别、操作第三方系统等）。
- 当用户明确点名某个工具或某类能力时。
- 当任务涉及具体时间区间、对象范围、过滤条件等，并且需要依赖工具才能得到准确结果时。

2. 工具命名与选择
- 每个工具在调用时的名称格式为：{serverName}__{toolName}。
  例如：服务器名为 "calendar"，工具名为 "create_event"，实际调用名为 "calendar__create_event"。
- 在选择工具时：
  - 先根据工具描述判断是否符合用户意图；
  - 如有多个类似工具，优先选择描述更精确、与当前场景更贴近的工具。

3. 构造工具调用参数
- 在调用工具前，应先向用户澄清目标、范围和约束（如时间区间、数量限制、过滤条件等）。
- 构造参数时：
  - 必须严格遵守工具参数的 JSON Schema 要求；
  - 不要省略必填字段，不要编造关键参数；
  - 对不确定的信息应先向用户确认，再发起调用。

4. 处理工具返回结果
- 若工具返回错误或空结果：
  - 根据返回信息解释可能原因，不要编造结果；
  - 必要时建议用户调整请求或参数。

下面是你当前可以使用的 MCP 工具列表，请在需要时主动选择合适的工具进行调用：
`

	var toolsDescBuilder strings.Builder
	toolsDescBuilder.WriteString("\n\n## 可用工具列表\n\n")

	for serverName, tools := range allTools {
		fmt.Fprintf(&toolsDescBuilder, "### 来自 %s 的工具：\n\n", serverName)
		for _, tool := range tools {
			fmt.Fprintf(&toolsDescBuilder, "- **%s**: %s\n", tool.Name, tool.Description)
		}
		toolsDescBuilder.WriteString("\n")
	}

	toolsDescBuilder.WriteString("调用工具时，请根据上述规则谨慎选择工具并构造参数。\n")

	return basePrompt + intro + toolsDescBuilder.String(), nil
}

// ExecuteToolCall 执行OpenAI函数调用
func (m *MCPManager) ExecuteToolCall(ctx context.Context, robotCtx robotctx.RobotContext, toolCall openai.ToolCall) (string, bool, error) {
	// 解析工具名称，提取服务器名称和原始工具名称
	serverName, toolName, err := m.parseToolName(toolCall.Function.Name)
	if err != nil {
		return "", false, err
	}

	// 解析参数
	var args map[string]any
	if toolCall.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			return "", false, fmt.Errorf("failed to parse tool arguments: %w", err)
		}
	}

	// 将 RobotContext 转换为 Meta（map[string]any）
	metaBytes, err := json.Marshal(robotCtx)
	if err != nil {
		return "", false, fmt.Errorf("failed to marshal robot context: %w", err)
	}
	var meta map[string]any
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return "", false, fmt.Errorf("failed to unmarshal robot context to meta: %w", err)
	}

	// 构建MCP调用参数
	params := &sdkmcp.CallToolParams{Meta: meta, Name: toolName, Arguments: args}

	// 调用MCP工具
	result, err := m.CallToolByName(ctx, serverName, params)
	if err != nil {
		return "", false, fmt.Errorf("failed to call mcp tool: %w", err)
	}

	return m.formatToolResult(result)
}

// parseToolName 解析工具名称
func (m *MCPManager) parseToolName(fullName string) (serverName, toolName string, err error) {
	// 工具名称格式：serverName__toolName
	for _, client := range m.GetAllClients() {
		prefix := client.GetConfig().Name + "__"
		if len(fullName) > len(prefix) && fullName[:len(prefix)] == prefix {
			return client.GetConfig().Name, fullName[len(prefix):], nil
		}
	}

	return "", "", fmt.Errorf("invalid tool name format: %s", fullName)
}

func (m *MCPManager) formatToolResult(result *sdkmcp.CallToolResult) (string, bool, error) {
	if result.IsError {
		if len(result.Content) > 0 {
			var errmsgs []string
			for _, content := range result.Content {
				if textContent, ok := content.(*sdkmcp.TextContent); ok {
					if textContent.Text != "" {
						errmsgs = append(errmsgs, textContent.Text)
					}
				}
			}
			if len(errmsgs) > 0 {
				return "", false, fmt.Errorf("MCP调用失败: %s", strings.Join(errmsgs, "\n"))
			}
		}
		return "", false, fmt.Errorf("MCP调用失败")
	}
	if result.StructuredContent != nil {
		type CallToolResult struct {
			IsCallToolResult bool `json:"is_call_tool_result,omitempty" jsonschema:"是否为调用工具结果"`
		}
		var callToolResult CallToolResult
		sb, err := json.Marshal(result.StructuredContent)
		if err != nil {
			return "", false, err
		}
		if err := json.Unmarshal(sb, &callToolResult); err != nil {
			return "", false, err
		}
		if callToolResult.IsCallToolResult {
			return string(sb), true, nil
		}
	}
	// 直接将结果序列化为字符串返回，交由上层决定发送策略
	rb, err := json.Marshal(result)
	if err != nil {
		return "", false, err
	}
	return string(rb), false, nil
}

// AddServer 添加并连接一个MCP服务器
func (m *MCPManager) AddServer(server *model.MCPServer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在
	if _, exists := m.clients[server.ID]; exists {
		return fmt.Errorf("mcp server %d already exists", server.ID)
	}

	// 创建客户端
	client, err := NewMCPClient(server)
	if err != nil {
		return fmt.Errorf("failed to create mcp client: %w", err)
	}

	// 连接到服务器
	// 使用管理器长期上下文，避免会话绑定 ctx 被提前取消
	if err := client.Connect(m.ctx); err != nil {
		return fmt.Errorf("failed to connect to mcp server: %w", err)
	}

	// 初始化MCP会话
	initCtx, initCancel := context.WithTimeout(context.Background(), time.Duration(server.ConnectTimeout)*time.Second)
	defer initCancel()
	serverInfo, err := client.Initialize(initCtx)
	if err != nil {
		client.Disconnect()
		return fmt.Errorf("failed to initialize mcp session: %w", err)
	}

	// 保存客户端
	m.clients[server.ID] = client

	// 更新数据库状态
	m.repo.UpdateConnectionSuccess(server.ID)

	log.Printf("MCP server '%s' (%s) connected successfully - Version: %s",
		server.Name, server.Transport, serverInfo.Version)

	// 启动心跳检测（如果启用）
	if server.HeartbeatEnable != nil && *server.HeartbeatEnable && server.HeartbeatInterval > 0 {
		heartbeatCtx, heartbeatCancel := context.WithCancel(m.ctx)
		m.heartbeatCancels[server.ID] = heartbeatCancel
		go m.startHeartbeat(heartbeatCtx, server, client, time.Duration(server.HeartbeatInterval)*time.Second)
	}

	return nil
}

// RemoveServer 移除并断开一个MCP服务器
func (m *MCPManager) RemoveServer(serverID uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[serverID]
	if !exists {
		return fmt.Errorf("mcp server %d not found", serverID)
	}

	// 停止心跳检测
	if cancel, exists := m.heartbeatCancels[serverID]; exists {
		cancel()
		delete(m.heartbeatCancels, serverID)
		log.Printf("MCP server %d heartbeat stopped", serverID)
	}

	// 断开连接
	if err := client.Disconnect(); err != nil {
		return fmt.Errorf("failed to disconnect mcp server: %w", err)
	}

	// 从映射中移除
	delete(m.clients, serverID)

	log.Printf("MCP server %d removed", serverID)
	return nil
}

// ReloadServer 重新加载一个MCP服务器（先断开再重连）
func (m *MCPManager) ReloadServer(serverID uint64) error {
	// 先移除
	if err := m.RemoveServer(serverID); err != nil && err.Error() != fmt.Sprintf("mcp server %d not found", serverID) {
		return err
	}

	// 从数据库重新加载配置
	server, err := m.repo.FindByID(serverID)
	if err != nil {
		return fmt.Errorf("failed to load server config: %w", err)
	}

	if server == nil {
		return fmt.Errorf("server %d not found in database", serverID)
	}

	// 检查是否启用
	if server.Enabled == nil || !*server.Enabled {
		return fmt.Errorf("server %d is not enabled", serverID)
	}

	// 重新添加
	return m.AddServer(server)
}

// GetClient 获取MCP客户端
func (m *MCPManager) GetClient(serverID uint64) (MCPClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[serverID]
	if !exists {
		return nil, fmt.Errorf("mcp server %d not found", serverID)
	}

	return client, nil
}

// GetClientByName 根据名称获取MCP客户端
func (m *MCPManager) GetClientByName(name string) (MCPClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.clients {
		if client.GetConfig().Name == name {
			return client, nil
		}
	}

	return nil, fmt.Errorf("mcp server '%s' not found", name)
}

// GetAllClients 获取所有MCP客户端
func (m *MCPManager) GetAllClients() []MCPClient {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients := make([]MCPClient, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}

	return clients
}

// GetAllTools 获取所有MCP服务器的工具列表
func (m *MCPManager) GetAllTools(ctx context.Context) (map[string][]*sdkmcp.Tool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allTools := make(map[string][]*sdkmcp.Tool)

	for _, client := range m.clients {
		tools, err := client.ListTools(ctx)
		if err != nil {
			log.Printf("Failed to list tools from server %s: %v",
				client.GetConfig().Name, err)
			continue
		}

		if len(tools) > 0 {
			allTools[client.GetConfig().Name] = tools
		}
	}

	return allTools, nil
}

// CallTool 调用指定服务器的工具
func (m *MCPManager) CallTool(ctx context.Context, serverID uint64, params *sdkmcp.CallToolParams) (*sdkmcp.CallToolResult, error) {
	client, err := m.GetClient(serverID)
	if err != nil {
		return nil, err
	}

	result, err := client.CallTool(ctx, params)
	if err != nil {
		// 记录错误
		m.repo.IncrementErrorCount(serverID)
		m.repo.UpdateConnectionError(serverID, err.Error())
		return nil, err
	}

	return result, nil
}

// CallToolByName 根据服务器名称调用工具
func (m *MCPManager) CallToolByName(ctx context.Context, serverName string, params *sdkmcp.CallToolParams) (*sdkmcp.CallToolResult, error) {
	client, err := m.GetClientByName(serverName)
	if err != nil {
		return nil, err
	}

	result, err := client.CallTool(ctx, params)
	if err != nil {
		// 记录错误
		serverID := client.GetConfig().ID
		m.repo.IncrementErrorCount(serverID)
		m.repo.UpdateConnectionError(serverID, err.Error())
		return nil, err
	}

	return result, nil
}

// GetActiveServerCount 获取活动服务器数量
func (m *MCPManager) GetActiveServerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// GetServerStats 获取服务器统计信息
func (m *MCPManager) GetServerStats(serverID uint64) (*MCPConnectionStats, error) {
	client, err := m.GetClient(serverID)
	if err != nil {
		return nil, err
	}

	return client.GetStats(), nil
}

// Shutdown 关闭管理器，断开所有连接
func (m *MCPManager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("Shutting down MCP Manager...")

	// 取消所有心跳检测
	for id, cancel := range m.heartbeatCancels {
		log.Printf("Stopping heartbeat for MCP server %d", id)
		cancel()
	}

	// 取消context
	m.cancelFunc()

	// 断开所有连接
	var lastErr error
	for id, client := range m.clients {
		if err := client.Disconnect(); err != nil {
			log.Printf("Failed to disconnect MCP server %d: %v", id, err)
			lastErr = err
		}
	}

	// 清空映射
	m.clients = make(map[uint64]MCPClient)
	m.heartbeatCancels = make(map[uint64]context.CancelFunc)

	log.Printf("MCP Manager shutdown complete")
	return lastErr
}

// startHeartbeat 启动心跳检测
func (m *MCPManager) startHeartbeat(ctx context.Context, server *model.MCPServer, client MCPClient, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	tempMaxRetry := 5
	tempRetryCount := 0

	log.Printf("MCP server %s heartbeat started (interval: %v)", server.Name, interval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("MCP server %s heartbeat context cancelled", server.Name)
			return
		case <-ticker.C:
			if !client.IsConnected() {
				log.Printf("MCP server %s client disconnected, stopping heartbeat", server.Name)
				return
			}

			pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := client.Ping(pingCtx)
			cancel()
			if err != nil {
				log.Printf("Heartbeat failed for MCP server %s: %v", server.Name, err)
				m.repo.IncrementErrorCount(server.ID)
				m.repo.UpdateConnectionError(server.ID, err.Error())

				// 忽略临时性关闭/重连中的错误，下轮再试
				if errors.Is(err, context.Canceled) ||
					strings.Contains(err.Error(), "client is closing") ||
					strings.Contains(err.Error(), "reconnect") {
					log.Printf("MCP %s heartbeat transient issue: %v", server.Name, err)
					tempRetryCount++
					if tempRetryCount < tempMaxRetry {
						continue
					}
				}

				go func(sid uint64) {
					if err := m.ReloadServer(sid); err != nil {
						log.Printf("Failed to reconnect MCP server %d: %v", sid, err)
					}
				}(server.ID)

				log.Printf("MCP server %s heartbeat stopped, waiting for reconnection", server.Name)
				return
			}

			tempRetryCount = 0
		}
	}
}

// EnableServer 启用服务器
func (m *MCPManager) EnableServer(serverID uint64) error {
	server, err := m.repo.FindByID(serverID)
	if err != nil {
		return err
	}
	err = m.AddServer(server)
	if err != nil {
		return err
	}
	if err := m.repo.UpdateEnabled(serverID, true); err != nil {
		return err
	}
	return nil
}

// DisableServer 禁用服务器
func (m *MCPManager) DisableServer(serverID uint64) error {
	if err := m.RemoveServer(serverID); err != nil {
		if err.Error() != fmt.Sprintf("mcp server %d not found", serverID) {
			return err
		}
	}
	return m.repo.UpdateEnabled(serverID, false)
}
