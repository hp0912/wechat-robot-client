package mcp

import (
	"time"
)

// 仅保留客户端内部用的统计与上下文字段

// MCPServerInfo MCP服务器信息（本地缓存使用）
type MCPServerInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Capabilities MCPCapabilities   `json:"capabilities"`
	Instructions string            `json:"instructions,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// MCPCapabilities MCP服务器能力（本地缓存使用）
type MCPCapabilities struct {
	Tools     bool `json:"tools"`
	Resources bool `json:"resources"`
	Prompts   bool `json:"prompts"`
}

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

// RobotContext 在工具调用入参中透传的机器人上下文
type RobotContext struct {
	WeChatClientPort string
	RobotID          int64
	RobotCode        string
	RobotRedisDB     uint
	RobotWxID        string
	FromWxID         string
	SenderWxID       string
	MessageID        int64
	RefMessageID     int64
}

// MessageSender 发送微信消息的适配器
type MessageSender interface {
	SendTextMessage(toWxID, content string, at ...string) error
	SendAppMessage(toWxID string, appMsgType int, appMsgXml string) error
}
