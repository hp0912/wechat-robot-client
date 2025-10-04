package mcp

import (
	"time"
	"wechat-robot-client/model"
)

const MCPProtocolVersion = "0.1.0"

// MCPRequest MCP请求基础结构
type MCPRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      string `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// MCPResponse MCP响应基础结构
type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      string    `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
}

// MCPError MCP错误结构
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MCPTool MCP工具定义
type MCPTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// MCPToolCallRequest 工具调用请求
type MCPToolCallRequest struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// MCPToolCallResponse 工具调用响应
type MCPToolCallResponse struct {
	Content any  `json:"content"`
	IsError bool `json:"isError"`
}

// MCPServerInfo MCP服务器信息
type MCPServerInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Capabilities MCPCapabilities   `json:"capabilities"`
	Instructions string            `json:"instructions,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// MCPCapabilities MCP服务器能力
type MCPCapabilities struct {
	Tools     bool `json:"tools"`
	Resources bool `json:"resources"`
	Prompts   bool `json:"prompts"`
}

// MCPResource MCP资源定义
type MCPResource struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	MimeType    string            `json:"mimeType,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// MCPPrompt MCP提示词定义
type MCPPrompt struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Arguments   []MCPPromptArgument `json:"arguments,omitempty"`
	Messages    []MCPPromptMessage  `json:"messages,omitempty"`
	Metadata    map[string]any      `json:"metadata,omitempty"`
}

// MCPPromptArgument 提示词参数
type MCPPromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}

// MCPPromptMessage 提示词消息
type MCPPromptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MCPInitializeParams 初始化参数
type MCPInitializeParams struct {
	ProtocolVersion string          `json:"protocolVersion"`
	ClientInfo      MCPClientInfo   `json:"clientInfo"`
	Capabilities    MCPCapabilities `json:"capabilities"`
}

// MCPClientInfo 客户端信息
type MCPClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MCPListToolsResult 工具列表结果
type MCPListToolsResult struct {
	Tools []MCPTool `json:"tools"`
}

// MCPCallToolParams 调用工具参数
type MCPCallToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// MCPCallToolResult 调用工具结果
type MCPCallToolResult struct {
	Content []MCPToolContent `json:"content"`
	IsError bool             `json:"isError,omitempty"`
}

// MCPToolContent 工具内容
type MCPToolContent struct {
	Type     model.MessageType    `json:"type"`
	SubType  model.AppMessageType `json:"subType,omitempty"`
	Text     string               `json:"text,omitempty"`
	Mentions []string             `json:"mentions,omitempty"` // 只对文本消息有效
	Data     any                  `json:"data,omitempty"`
}

// MCPListResourcesResult 资源列表结果
type MCPListResourcesResult struct {
	Resources []MCPResource `json:"resources"`
}

// MCPReadResourceParams 读取资源参数
type MCPReadResourceParams struct {
	URI string `json:"uri"`
}

// MCPReadResourceResult 读取资源结果
type MCPReadResourceResult struct {
	Contents []MCPResourceContent `json:"contents"`
}

// MCPResourceContent 资源内容
type MCPResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
}

// MCPPingParams Ping参数
type MCPPingParams struct{}

// MCPPingResult Ping结果
type MCPPingResult struct{}

// MCPConnectionStats MCP连接统计
type MCPConnectionStats struct {
	ConnectedAt    time.Time     `json:"connectedAt"`
	LastActiveAt   time.Time     `json:"lastActiveAt"`
	RequestCount   int64         `json:"requestCount"`
	SuccessCount   int64         `json:"successCount"`
	ErrorCount     int64         `json:"errorCount"`
	AverageLatency time.Duration `json:"averageLatency"`
	IsConnected    bool          `json:"isConnected"`
}

type RobotContext struct {
	// 微信机器人实例 ID
	RobotID      int64
	RobotCode    string
	RobotRedisDB uint
	// 微信机器人实例的微信 ID
	RobotWxID string
	// 消息来源，如果是群聊则为群ID，私聊则为用户ID
	FromWxID string
	// 发送者ID，群聊中为发送者的用户ID，私聊中同FromWxID
	SenderWxID string
	// 当前消息 ID
	MessageID int64
	// 引用消息 ID
	RefMessageID int64
}

type MessageSender interface {
	SendTextMessage(toWxID, content string, at ...string) error
	SendAppMessage(toWxID string, appMsgType int, appMsgXml string) error
}
