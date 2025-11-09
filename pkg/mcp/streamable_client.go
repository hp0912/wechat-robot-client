package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"wechat-robot-client/model"
)

type StreamableClient struct {
	*BaseClient
	client     *sdkmcp.Client
	session    *sdkmcp.ClientSession
	url        string
	authHeader string
	headers    map[string]string
}

// NewStreamableClient 创建Streamable客户端
func NewStreamableClient(config *model.MCPServer) *StreamableClient {
	headers, _ := config.GetHeaders()
	authHeader := ""
	if config.NeedsAuth() {
		switch config.AuthType {
		case model.MCPAuthTypeBearer:
			authHeader = "Bearer " + config.AuthToken
		case model.MCPAuthTypeAPIKey:
			authHeader = "ApiKey " + config.AuthToken
		case model.MCPAuthTypeBasic:
			creds := config.AuthUsername + ":" + config.AuthPassword
			authHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))
		}
	}

	return &StreamableClient{
		BaseClient: NewBaseClient(config),
		url:        config.URL,
		authHeader: authHeader,
		headers:    headers,
	}
}

func (c *StreamableClient) Auth(h sdkmcp.MethodHandler) sdkmcp.MethodHandler {
	return func(ctx context.Context, method string, req sdkmcp.Request) (result sdkmcp.Result, err error) {
		header := req.GetExtra().Header
		if c.authHeader != "" {
			header.Set("Authorization", c.authHeader)
		}
		for k, v := range c.headers {
			if v != "" {
				header.Set(k, v)
			}
		}
		return h(ctx, method, req)
	}
}

// Connect 连接到MCP服务器
func (c *StreamableClient) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return ErrAlreadyConnected
	}

	transport := &sdkmcp.StreamableClientTransport{Endpoint: c.url}
	clientName := c.config.ClientName
	if clientName == "" {
		clientName = "wechat-robot-mcp-client"
	}
	c.client = sdkmcp.NewClient(&sdkmcp.Implementation{
		Name:    clientName,
		Version: "1.0.0",
	}, nil)
	c.client.AddSendingMiddleware(c.Auth)

	sess, err := c.client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to mcp server: %w", err)
	}

	c.session = sess
	c.setConnected(true)
	return nil
}

// Disconnect 断开连接
func (c *StreamableClient) Disconnect() error {
	if !c.IsConnected() {
		return nil
	}

	if c.session != nil {
		c.session.Close()
		c.session = nil
	}

	c.setConnected(false)
	return nil
}

// Initialize 初始化MCP会话
func (c *StreamableClient) Initialize(ctx context.Context) (*MCPServerInfo, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	// 通过探测接口推断能力
	cap := MCPCapabilities{}
	if _, err := c.session.ListTools(ctx, &sdkmcp.ListToolsParams{}); err == nil {
		cap.Tools = true
	}
	if _, err := c.session.ListResources(ctx, &sdkmcp.ListResourcesParams{}); err == nil {
		cap.Resources = true
	}
	if _, err := c.session.ListPrompts(ctx, &sdkmcp.ListPromptsParams{}); err == nil {
		cap.Prompts = true
	}

	info := &MCPServerInfo{
		Name:         c.config.Name,
		Version:      "1.0.0",
		Capabilities: cap,
	}
	c.setServerInfo(info)
	return info, nil
}

// ListTools 列出所有可用工具
func (c *StreamableClient) ListTools(ctx context.Context) ([]*sdkmcp.Tool, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	start := time.Now()
	toolsRes, err := c.session.ListTools(ctx, &sdkmcp.ListToolsParams{})
	c.updateStats(err == nil, time.Since(start))
	if err != nil {
		return nil, err
	}
	return toolsRes.Tools, nil
}

// CallTool 调用工具
func (c *StreamableClient) CallTool(ctx context.Context, params *sdkmcp.CallToolParams) (*sdkmcp.CallToolResult, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	start := time.Now()
	res, err := c.session.CallTool(ctx, params)
	c.updateStats(err == nil, time.Since(start))

	if err != nil {
		return nil, err
	}
	return res, nil
}

// ListResources 列出所有可用资源
func (c *StreamableClient) ListResources(ctx context.Context) ([]*sdkmcp.Resource, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	start := time.Now()
	items, err := c.session.ListResources(ctx, &sdkmcp.ListResourcesParams{})
	c.updateStats(err == nil, time.Since(start))

	if err != nil {
		return nil, err
	}
	return items.Resources, nil
}

// ReadResource 读取资源
func (c *StreamableClient) ReadResource(ctx context.Context, params *sdkmcp.ReadResourceParams) (*sdkmcp.ReadResourceResult, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	start := time.Now()
	rr, err := c.session.ReadResource(ctx, params)
	c.updateStats(err == nil, time.Since(start))

	if err != nil {
		return nil, err
	}
	return rr, nil
}

// Ping 心跳检测
func (c *StreamableClient) Ping(ctx context.Context) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	start := time.Now()
	err := c.session.Ping(ctx, &sdkmcp.PingParams{})
	c.updateStats(err == nil, time.Since(start))

	return err
}
