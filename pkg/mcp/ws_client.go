package mcp

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"wechat-robot-client/model"

	"github.com/gorilla/websocket"
)

// WebSocketClient WebSocket模式的MCP客户端
type WebSocketClient struct {
	*BaseClient
	conn              *websocket.Conn
	mu                sync.Mutex
	pendingRequests   map[string]chan *MCPResponse
	pendingMu         sync.RWMutex
	heartbeatTicker   *time.Ticker
	heartbeatStopChan chan struct{}
}

// NewWebSocketClient 创建WebSocket客户端
func NewWebSocketClient(config *model.MCPServer) *WebSocketClient {
	return &WebSocketClient{
		BaseClient:        NewBaseClient(config),
		pendingRequests:   make(map[string]chan *MCPResponse),
		heartbeatStopChan: make(chan struct{}),
	}
}

// Connect 连接到MCP服务器
func (c *WebSocketClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsConnected() {
		return ErrAlreadyConnected
	}

	// 创建WebSocket拨号器
	dialer := websocket.Dialer{
		HandshakeTimeout: c.getConnectTimeout(),
	}

	// 配置TLS
	if c.config.TLSSkipVerify != nil && *c.config.TLSSkipVerify {
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// 创建请求头
	header := http.Header{}
	header.Set("Content-Type", "application/json")

	// 解析自定义请求头
	if c.config.Headers != nil {
		var headers map[string]string
		if err := json.Unmarshal(c.config.Headers, &headers); err == nil {
			for k, v := range headers {
				header.Set(k, v)
			}
		}
	}

	// 添加认证
	if err := c.addAuthentication(header); err != nil {
		return err
	}

	// 连接WebSocket
	conn, _, err := dialer.DialContext(ctx, c.config.URL, header)
	if err != nil {
		return fmt.Errorf("failed to dial websocket: %w", err)
	}

	c.conn = conn
	c.setConnected(true)

	// 启动消息读取协程
	go c.readMessages()

	// 启动心跳检测
	if c.config.HeartbeatEnable != nil && *c.config.HeartbeatEnable {
		c.startHeartbeat()
	}

	return nil
}

// Disconnect 断开连接
func (c *WebSocketClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.IsConnected() {
		return nil
	}

	c.setConnected(false)

	// 停止心跳
	c.stopHeartbeat()

	// 关闭所有待处理的请求
	c.pendingMu.Lock()
	for _, ch := range c.pendingRequests {
		close(ch)
	}
	c.pendingRequests = make(map[string]chan *MCPResponse)
	c.pendingMu.Unlock()

	// 关闭WebSocket连接
	if c.conn != nil {
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		c.conn = nil
	}

	return nil
}

// Initialize 初始化MCP会话
func (c *WebSocketClient) Initialize(ctx context.Context) (*MCPServerInfo, error) {
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
func (c *WebSocketClient) ListTools(ctx context.Context) ([]MCPTool, error) {
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
func (c *WebSocketClient) CallTool(ctx context.Context, params MCPCallToolParams) (*MCPCallToolResult, error) {
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
func (c *WebSocketClient) ListResources(ctx context.Context) ([]MCPResource, error) {
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
func (c *WebSocketClient) ReadResource(ctx context.Context, params MCPReadResourceParams) (*MCPReadResourceResult, error) {
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
func (c *WebSocketClient) Ping(ctx context.Context) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	var result MCPPingResult
	return c.sendRequest(ctx, "ping", MCPPingParams{}, &result)
}

// sendRequest 发送请求
func (c *WebSocketClient) sendRequest(ctx context.Context, method string, params any, result any) error {
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

	// 发送请求
	c.mu.Lock()
	if err := c.conn.WriteMessage(websocket.TextMessage, reqData); err != nil {
		c.mu.Unlock()
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to send websocket message: %w", err)
	}
	c.mu.Unlock()

	// 等待响应
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

// readMessages 读取WebSocket消息
func (c *WebSocketClient) readMessages() {
	for c.IsConnected() {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if c.IsConnected() {
				// 连接异常断开
				c.Disconnect()
			}
			return
		}

		// 解析响应
		var resp MCPResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			continue
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
}

// startHeartbeat 启动心跳检测
func (c *WebSocketClient) startHeartbeat() {
	interval := time.Duration(c.config.HeartbeatInterval) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}

	c.heartbeatTicker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-c.heartbeatTicker.C:
				if err := c.Ping(context.Background()); err != nil {
					// 心跳失败，断开连接
					c.Disconnect()
					return
				}
			case <-c.heartbeatStopChan:
				return
			}
		}
	}()
}

// stopHeartbeat 停止心跳检测
func (c *WebSocketClient) stopHeartbeat() {
	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
		close(c.heartbeatStopChan)
		c.heartbeatStopChan = make(chan struct{})
	}
}

// addAuthentication 添加认证信息到请求头
func (c *WebSocketClient) addAuthentication(header http.Header) error {
	switch c.config.AuthType {
	case model.MCPAuthTypeNone:
		return nil

	case model.MCPAuthTypeBearer:
		if c.config.AuthToken == "" {
			return ErrAuthenticationFailed
		}
		header.Set("Authorization", "Bearer "+c.config.AuthToken)
		return nil

	case model.MCPAuthTypeBasic:
		if c.config.AuthUsername == "" || c.config.AuthPassword == "" {
			return ErrAuthenticationFailed
		}
		auth := c.config.AuthUsername + ":" + c.config.AuthPassword
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
		header.Set("Authorization", "Basic "+encodedAuth)
		return nil

	case model.MCPAuthTypeAPIKey:
		if c.config.AuthToken == "" {
			return ErrAuthenticationFailed
		}
		header.Set("X-API-Key", c.config.AuthToken)
		return nil

	default:
		return fmt.Errorf("unsupported auth type: %s", c.config.AuthType)
	}
}
