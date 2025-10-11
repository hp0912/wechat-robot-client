package mcp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"wechat-robot-client/model"
	"wechat-robot-client/repository"

	"gorm.io/gorm"
)

// MCPManager MCP服务管理器
type MCPManager struct {
	db         *gorm.DB
	clients    map[uint64]MCPClient
	mu         sync.RWMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
	repo       *repository.MCPServer
}

// NewMCPManager 创建MCP管理器
func NewMCPManager(db *gorm.DB) *MCPManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &MCPManager{
		db:         db,
		clients:    make(map[uint64]MCPClient),
		ctx:        ctx,
		cancelFunc: cancel,
		repo:       repository.NewMCPServerRepo(ctx, db),
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
	ctx, cancel := context.WithTimeout(m.ctx, time.Duration(server.ConnectTimeout)*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to mcp server: %w", err)
	}

	// 初始化MCP会话
	serverInfo, err := client.Initialize(ctx)
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
	if server.HeartbeatEnable != nil && *server.HeartbeatEnable {
		go m.startHeartbeat(server.ID, client, time.Duration(server.HeartbeatInterval)*time.Second)
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
func (m *MCPManager) GetAllTools(ctx context.Context) (map[string][]MCPTool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allTools := make(map[string][]MCPTool)

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
func (m *MCPManager) CallTool(ctx context.Context, serverID uint64, params MCPCallToolParams) (*MCPCallToolResult, error) {
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
func (m *MCPManager) CallToolByName(ctx context.Context, serverName string, params MCPCallToolParams) (*MCPCallToolResult, error) {
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

	log.Printf("MCP Manager shutdown complete")
	return lastErr
}

// startHeartbeat 启动心跳检测
func (m *MCPManager) startHeartbeat(serverID uint64, client MCPClient, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if !client.IsConnected() {
				return
			}

			ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
			err := client.Ping(ctx)
			cancel()

			if err != nil {
				log.Printf("Heartbeat failed for MCP server %d: %v", serverID, err)
				m.repo.IncrementErrorCount(serverID)
				m.repo.UpdateConnectionError(serverID, err.Error())

				// 尝试重连
				if err := m.ReloadServer(serverID); err != nil {
					log.Printf("Failed to reconnect MCP server %d: %v", serverID, err)
				}
			}
		}
	}
}

// EnableServer 启用服务器
func (m *MCPManager) EnableServer(serverID uint64) error {
	// 更新数据库
	if err := m.repo.UpdateEnabled(serverID, true); err != nil {
		return err
	}

	// 加载服务器
	server, err := m.repo.FindByID(serverID)
	if err != nil {
		return err
	}

	return m.AddServer(server)
}

// DisableServer 禁用服务器
func (m *MCPManager) DisableServer(serverID uint64) error {
	// 移除服务器
	if err := m.RemoveServer(serverID); err != nil {
		// 如果服务器不存在，忽略错误
		if err.Error() != fmt.Sprintf("mcp server %d not found", serverID) {
			return err
		}
	}

	// 更新数据库
	return m.repo.UpdateEnabled(serverID, false)
}
