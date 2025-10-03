package mcp

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"wechat-robot-client/model"
)

// SSEClient SSE模式的MCP客户端
type SSEClient struct {
	*BaseClient
	httpClient      *http.Client
	conn            io.ReadCloser
	mu              sync.Mutex
	pendingRequests map[string]chan *MCPResponse
	pendingMu       sync.RWMutex
	cancelFunc      context.CancelFunc
}

// NewSSEClient 创建SSE客户端
func NewSSEClient(config *model.MCPServer) *SSEClient {
	// 创建HTTP客户端
	transport := &http.Transport{}

	// 配置TLS
	if config.TLSSkipVerify != nil && *config.TLSSkipVerify {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   0, // SSE需要长连接，不设置超时
	}

	return &SSEClient{
		BaseClient:      NewBaseClient(config),
		httpClient:      httpClient,
		pendingRequests: make(map[string]chan *MCPResponse),
	}
}

// Connect 连接到MCP服务器
func (c *SSEClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsConnected() {
		return ErrAlreadyConnected
	}

	// 创建可取消的context
	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel

	// 创建SSE请求
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.URL, nil)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create sse request: %w", err)
	}

	// 设置SSE请求头
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// 解析自定义请求头
	if c.config.Headers != nil {
		var headers map[string]string
		if err := json.Unmarshal(c.config.Headers, &headers); err == nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
	}

	// 添加认证
	if err := c.addAuthentication(req); err != nil {
		cancel()
		return err
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to connect sse: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		cancel()
		resp.Body.Close()
		return fmt.Errorf("sse connection failed with status: %d", resp.StatusCode)
	}

	c.conn = resp.Body
	c.setConnected(true)

	// 启动消息读取协程
	go c.readMessages()

	return nil
}

// Disconnect 断开连接
func (c *SSEClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.IsConnected() {
		return nil
	}

	c.setConnected(false)

	// 取消context
	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	// 关闭所有待处理的请求
	c.pendingMu.Lock()
	for _, ch := range c.pendingRequests {
		close(ch)
	}
	c.pendingRequests = make(map[string]chan *MCPResponse)
	c.pendingMu.Unlock()

	// 关闭连接
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	return nil
}

// Initialize 初始化MCP会话
func (c *SSEClient) Initialize(ctx context.Context) (*MCPServerInfo, error) {
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
func (c *SSEClient) ListTools(ctx context.Context) ([]MCPTool, error) {
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
func (c *SSEClient) CallTool(ctx context.Context, params MCPCallToolParams) (*MCPCallToolResult, error) {
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
func (c *SSEClient) ListResources(ctx context.Context) ([]MCPResource, error) {
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
func (c *SSEClient) ReadResource(ctx context.Context, params MCPReadResourceParams) (*MCPReadResourceResult, error) {
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
func (c *SSEClient) Ping(ctx context.Context) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	// SSE模式下，连接本身就是长连接，不需要额外的ping
	// 这里简单返回成功即可
	return nil
}

// sendRequest 发送请求（SSE模式下需要通过其他方式发送请求，这里简化处理）
// 注意：SSE是单向流，如果需要发送请求，需要配合其他通道（如HTTP POST）
func (c *SSEClient) sendRequest(ctx context.Context, method string, params any, result any) error {
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

	// 创建响应通道
	respChan := make(chan *MCPResponse, 1)
	c.pendingMu.Lock()
	c.pendingRequests[req.ID] = respChan
	c.pendingMu.Unlock()

	// 确保清理
	defer func() {
		c.pendingMu.Lock()
		delete(c.pendingRequests, req.ID)
		c.pendingMu.Unlock()
	}()

	// SSE模式下，通过HTTP POST发送请求到控制端点
	// 这里假设控制端点为 ${URL}/control
	controlURL := strings.TrimSuffix(c.config.URL, "/") + "/control"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", controlURL, strings.NewReader(string(reqData)))
	if err != nil {
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to create control request: %w", err)
	}

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

	if err := c.addAuthentication(httpReq); err != nil {
		c.updateStats(false, time.Since(start))
		return err
	}

	// 发送请求
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to send control request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusAccepted {
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("control request failed with status: %d", httpResp.StatusCode)
	}

	// 等待响应（通过SSE流接收）
	select {
	case resp := <-respChan:
		if resp == nil {
			c.updateStats(false, time.Since(start))
			return ErrConnectionClosed
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

	case <-ctx.Done():
		c.updateStats(false, time.Since(start))
		return ctx.Err()

	case <-time.After(c.getReadTimeout()):
		c.updateStats(false, time.Since(start))
		return ErrTimeout
	}
}

// readMessages 读取SSE消息
func (c *SSEClient) readMessages() {
	reader := bufio.NewReader(c.conn)
	var eventData strings.Builder

	for c.IsConnected() {
		line, err := reader.ReadString('\n')
		if err != nil {
			if c.IsConnected() {
				// 连接异常断开
				c.Disconnect()
			}
			return
		}

		line = strings.TrimSpace(line)

		// 空行表示事件结束
		if line == "" {
			if eventData.Len() > 0 {
				c.handleEvent(eventData.String())
				eventData.Reset()
			}
			continue
		}

		// 解析SSE数据行
		data, _ := strings.CutPrefix(line, "data: ")
		eventData.WriteString(data)
		// 忽略其他SSE字段（id:, event:, retry:）
	}
}

// handleEvent 处理SSE事件
func (c *SSEClient) handleEvent(data string) {
	// 解析响应
	var resp MCPResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		return
	}

	// 分发响应到对应的请求
	c.pendingMu.RLock()
	respChan, exists := c.pendingRequests[resp.ID]
	c.pendingMu.RUnlock()

	if exists {
		select {
		case respChan <- &resp:
		default:
			// 通道已满，丢弃响应
		}
	}
}

// addAuthentication 添加认证信息
func (c *SSEClient) addAuthentication(req *http.Request) error {
	switch c.config.AuthType {
	case model.MCPAuthTypeNone:
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
