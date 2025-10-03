package mcp

import (
	"context"
	"sync/atomic"
	"time"

	"wechat-robot-client/model"
)

// MCPClient MCP客户端接口
type MCPClient interface {
	// Connect 连接到MCP服务器
	Connect(ctx context.Context) error

	// Disconnect 断开连接
	Disconnect() error

	// IsConnected 检查是否已连接
	IsConnected() bool

	// Initialize 初始化MCP会话
	Initialize(ctx context.Context) (*MCPServerInfo, error)

	// ListTools 列出所有可用工具
	ListTools(ctx context.Context) ([]MCPTool, error)

	// CallTool 调用工具
	CallTool(ctx context.Context, params MCPCallToolParams) (*MCPCallToolResult, error)

	// ListResources 列出所有可用资源
	ListResources(ctx context.Context) ([]MCPResource, error)

	// ReadResource 读取资源
	ReadResource(ctx context.Context, params MCPReadResourceParams) (*MCPReadResourceResult, error)

	// Ping 心跳检测
	Ping(ctx context.Context) error

	// GetServerInfo 获取服务器信息
	GetServerInfo() *MCPServerInfo

	// GetStats 获取连接统计
	GetStats() *MCPConnectionStats

	// GetConfig 获取服务器配置
	GetConfig() *model.MCPServer
}

// BaseClient MCP客户端基础实现
type BaseClient struct {
	config     *model.MCPServer
	serverInfo *MCPServerInfo
	connected  atomic.Bool
	stats      MCPConnectionStats
	requestID  atomic.Int64
}

// NewBaseClient 创建基础客户端
func NewBaseClient(config *model.MCPServer) *BaseClient {
	return &BaseClient{
		config: config,
		stats: MCPConnectionStats{
			IsConnected: false,
		},
	}
}

// IsConnected 检查是否已连接
func (c *BaseClient) IsConnected() bool {
	return c.connected.Load()
}

// GetServerInfo 获取服务器信息
func (c *BaseClient) GetServerInfo() *MCPServerInfo {
	return c.serverInfo
}

// GetStats 获取连接统计
func (c *BaseClient) GetStats() *MCPConnectionStats {
	return &c.stats
}

// GetConfig 获取服务器配置
func (c *BaseClient) GetConfig() *model.MCPServer {
	return c.config
}

// setConnected 设置连接状态
func (c *BaseClient) setConnected(connected bool) {
	c.connected.Store(connected)
	c.stats.IsConnected = connected
	if connected {
		c.stats.ConnectedAt = time.Now()
		c.stats.LastActiveAt = time.Now()
	}
}

// setServerInfo 设置服务器信息
func (c *BaseClient) setServerInfo(info *MCPServerInfo) {
	c.serverInfo = info
}

// updateStats 更新统计信息
func (c *BaseClient) updateStats(success bool, latency time.Duration) {
	c.stats.LastActiveAt = time.Now()
	c.stats.RequestCount++
	if success {
		c.stats.SuccessCount++
	} else {
		c.stats.ErrorCount++
	}
	if c.stats.AverageLatency == 0 {
		c.stats.AverageLatency = latency
	} else {
		c.stats.AverageLatency = (c.stats.AverageLatency + latency) / 2
	}
}

// nextRequestID 生成下一个请求ID
func (c *BaseClient) nextRequestID() string {
	id := c.requestID.Add(1)
	return c.config.Name + "-" + string(rune(id))
}

// getTimeout 获取超时时间
func (c *BaseClient) getConnectTimeout() time.Duration {
	if c.config.ConnectTimeout > 0 {
		return time.Duration(c.config.ConnectTimeout) * time.Second
	}
	return 30 * time.Second
}

func (c *BaseClient) getReadTimeout() time.Duration {
	if c.config.ReadTimeout > 0 {
		return time.Duration(c.config.ReadTimeout) * time.Second
	}
	return 60 * time.Second
}

func (c *BaseClient) getWriteTimeout() time.Duration {
	if c.config.WriteTimeout > 0 {
		return time.Duration(c.config.WriteTimeout) * time.Second
	}
	return 60 * time.Second
}

// NewMCPClient 根据配置创建MCP客户端
func NewMCPClient(config *model.MCPServer) (MCPClient, error) {
	switch config.Transport {
	case model.MCPTransportTypeStdio:
		return NewStdioClient(config), nil
	case model.MCPTransportTypeHTTP:
		return NewHTTPClient(config), nil
	case model.MCPTransportTypeWS:
		return NewWebSocketClient(config), nil
	case model.MCPTransportTypeSSE:
		return NewSSEClient(config), nil
	default:
		return nil, ErrInvalidTransport
	}
}
