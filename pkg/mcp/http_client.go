package mcp

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"wechat-robot-client/model"
)

// HTTPClient HTTP模式的MCP客户端
type HTTPClient struct {
	*BaseClient
	httpClient *http.Client
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient(config *model.MCPServer) *HTTPClient {
	transport := &http.Transport{}
	if config.TLSSkipVerify != nil && *config.TLSSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(config.ReadTimeout) * time.Second,
	}
	return &HTTPClient{
		BaseClient: NewBaseClient(config),
		httpClient: httpClient,
	}
}

// Connect 连接到MCP服务器（HTTP模式只是验证连接）
func (c *HTTPClient) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return ErrAlreadyConnected
	}
	// HTTP模式下，连接即为成功（无需建立持久连接）
	c.setConnected(true)
	return nil
}

// Disconnect 断开连接
func (c *HTTPClient) Disconnect() error {
	if !c.IsConnected() {
		return nil
	}

	c.setConnected(false)
	c.httpClient.CloseIdleConnections()
	return nil
}

// Initialize 初始化MCP会话
func (c *HTTPClient) Initialize(ctx context.Context) (*MCPServerInfo, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	params := MCPInitializeParams{
		ProtocolVersion: MCPProtocolVersion,
		ClientInfo: MCPClientInfo{
			Name:    "wechat-robot-mcp-client",
			Version: "1.0.0",
		},
		Capabilities: MCPCapabilities{
			Tools:     true,
			Resources: true,
			Prompts:   true,
		},
	}

	var result MCPServerInfo
	if err := c.sendRequest(ctx, "initialize", params, &result); err != nil {
		return nil, err
	}

	c.setServerInfo(&result)
	return &result, nil
}

// ListTools 列出所有可用工具
func (c *HTTPClient) ListTools(ctx context.Context) ([]MCPTool, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	var result MCPListToolsResult
	if err := c.sendRequest(ctx, "tools/list", nil, &result); err != nil {
		return nil, err
	}

	return result.Tools, nil
}

// CallTool 调用工具
func (c *HTTPClient) CallTool(ctx context.Context, params MCPCallToolParams) (*MCPCallToolResult, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	var result MCPCallToolResult
	if err := c.sendRequest(ctx, "tools/call", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ListResources 列出所有可用资源
func (c *HTTPClient) ListResources(ctx context.Context) ([]MCPResource, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	var result MCPListResourcesResult
	if err := c.sendRequest(ctx, "resources/list", nil, &result); err != nil {
		return nil, err
	}

	return result.Resources, nil
}

// ReadResource 读取资源
func (c *HTTPClient) ReadResource(ctx context.Context, params MCPReadResourceParams) (*MCPReadResourceResult, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	var result MCPReadResourceResult
	if err := c.sendRequest(ctx, "resources/read", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Ping 心跳检测
func (c *HTTPClient) Ping(ctx context.Context) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	var result MCPPingResult
	return c.sendRequest(ctx, "ping", MCPPingParams{}, &result)
}

// sendRequest 发送HTTP请求
func (c *HTTPClient) sendRequest(ctx context.Context, method string, params any, result any) error {
	start := time.Now()

	// 创建请求
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  method,
		Params:  params,
	}

	// 序列化请求
	reqData, err := json.Marshal(req)
	if err != nil {
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.URL, bytes.NewReader(reqData))
	if err != nil {
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to create http request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")

	// 解析自定义请求头
	if c.config.Headers != nil {
		var headers map[string]string
		if err := json.Unmarshal(c.config.Headers, &headers); err == nil {
			for k, v := range headers {
				httpReq.Header.Set(k, v)
			}
		}
	}

	// 添加认证
	if err := c.addAuthentication(httpReq); err != nil {
		c.updateStats(false, time.Since(start))
		return err
	}

	// 发送请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to send http request: %w", err)
	}
	defer httpResp.Body.Close()

	// 检查HTTP状态码
	if httpResp.StatusCode != http.StatusOK {
		c.updateStats(false, time.Since(start))
		bodyBytes, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("http request failed with status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	// 读取响应
	respData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to read http response: %w", err)
	}

	// 解析响应
	var resp MCPResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 检查错误
	if resp.Error != nil {
		c.updateStats(false, time.Since(start))
		return resp.Error
	}

	// 解析结果
	if result != nil && resp.Result != nil {
		resultData, err := json.Marshal(resp.Result)
		if err != nil {
			c.updateStats(false, time.Since(start))
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		if err := json.Unmarshal(resultData, result); err != nil {
			c.updateStats(false, time.Since(start))
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	c.updateStats(true, time.Since(start))
	return nil
}

// addAuthentication 添加认证信息
func (c *HTTPClient) addAuthentication(req *http.Request) error {
	switch c.config.AuthType {
	case model.MCPAuthTypeNone:
		// 无需认证
		return nil

	case model.MCPAuthTypeBearer:
		if c.config.AuthToken == "" {
			return ErrAuthenticationFailed
		}
		req.Header.Set("Authorization", "Bearer "+c.config.AuthToken)
		return nil

	case model.MCPAuthTypeBasic:
		if c.config.AuthUsername == "" || c.config.AuthPassword == "" {
			return ErrAuthenticationFailed
		}
		auth := c.config.AuthUsername + ":" + c.config.AuthPassword
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Set("Authorization", "Basic "+encodedAuth)
		return nil

	case model.MCPAuthTypeAPIKey:
		if c.config.AuthToken == "" {
			return ErrAuthenticationFailed
		}
		req.Header.Set("X-API-Key", c.config.AuthToken)
		return nil

	default:
		return fmt.Errorf("unsupported auth type: %s", c.config.AuthType)
	}
}
