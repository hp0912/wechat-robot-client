package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
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

type authRoundTripper struct {
	base       http.RoundTripper
	authHeader string
	headers    map[string]string
}

func (t *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.authHeader != "" {
		req.Header.Set("Authorization", t.authHeader)
	}
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
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

// Connect 连接到MCP服务器
func (c *StreamableClient) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return ErrAlreadyConnected
	}

	httpClient := &http.Client{
		Transport: &authRoundTripper{
			base:       http.DefaultTransport,
			authHeader: c.authHeader,
			headers:    c.headers,
		},
	}

	transport := &sdkmcp.StreamableClientTransport{
		Endpoint:   c.url,
		HTTPClient: httpClient,
	}

	clientName := c.config.ClientName
	if clientName == "" {
		clientName = "wechat-robot-mcp-client"
	}
	c.client = sdkmcp.NewClient(&sdkmcp.Implementation{
		Name:    clientName,
		Version: "1.0.0",
	}, nil)

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
