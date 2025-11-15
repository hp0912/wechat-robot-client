package mcp

import (
	"errors"
	"fmt"
)

var (
	// ErrNotConnected MCP客户端未连接
	ErrNotConnected = errors.New("mcp client not connected")

	// ErrAlreadyConnected MCP客户端已连接
	ErrAlreadyConnected = errors.New("mcp client already connected")

	// ErrInvalidTransport 无效的传输类型
	ErrInvalidTransport = errors.New("invalid transport type")

	// ErrInvalidRequest 无效的请求
	ErrInvalidRequest = errors.New("invalid mcp request")

	// ErrInvalidResponse 无效的响应
	ErrInvalidResponse = errors.New("invalid mcp response")

	// ErrToolNotFound 工具未找到
	ErrToolNotFound = errors.New("tool not found")

	// ErrResourceNotFound 资源未找到
	ErrResourceNotFound = errors.New("resource not found")

	// ErrTimeout 请求超时
	ErrTimeout = errors.New("mcp request timeout")

	// ErrServerError 服务器错误
	ErrServerError = errors.New("mcp server error")

	// ErrAuthenticationFailed 认证失败
	ErrAuthenticationFailed = errors.New("authentication failed")

	// ErrConnectionClosed 连接已关闭
	ErrConnectionClosed = errors.New("connection closed")
)

// MCPErrorCode MCP错误码
const (
	// JSON-RPC标准错误码
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603

	// MCP自定义错误码
	ToolExecutionError  = -32000
	ResourceAccessError = -32001
	ConnectionError     = -32002
	AuthenticationError = -32003
	TimeoutError        = -32004
)

// NewMCPError 创建MCP错误（本地轻量错误结构，避免依赖已移除的协议镜像类型）
type MCPError struct {
	Code    int
	Message string
	Data    any
}

func NewMCPError(code int, message string, data any) *MCPError {
	return &MCPError{Code: code, Message: message, Data: data}
}

// Error 实现error接口
func (e *MCPError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("MCP Error [%d]: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("MCP Error [%d]: %s", e.Code, e.Message)
}

// WrapError 包装错误为MCP错误
func WrapError(err error) *MCPError {
	if err == nil {
		return nil
	}

	if mcpErr, ok := err.(*MCPError); ok {
		return mcpErr
	}

	switch {
	case errors.Is(err, ErrNotConnected), errors.Is(err, ErrConnectionClosed):
		return NewMCPError(ConnectionError, err.Error(), nil)
	case errors.Is(err, ErrAuthenticationFailed):
		return NewMCPError(AuthenticationError, err.Error(), nil)
	case errors.Is(err, ErrTimeout):
		return NewMCPError(TimeoutError, err.Error(), nil)
	case errors.Is(err, ErrToolNotFound), errors.Is(err, ErrResourceNotFound):
		return NewMCPError(MethodNotFound, err.Error(), nil)
	case errors.Is(err, ErrInvalidRequest):
		return NewMCPError(InvalidRequest, err.Error(), nil)
	default:
		return NewMCPError(InternalError, err.Error(), nil)
	}
}
