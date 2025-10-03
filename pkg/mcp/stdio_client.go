package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"wechat-robot-client/model"
)

// StdioClient Stdio模式的MCP客户端
type StdioClient struct {
	*BaseClient
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mu     sync.Mutex
}

// NewStdioClient 创建Stdio客户端
func NewStdioClient(config *model.MCPServer) *StdioClient {
	return &StdioClient{
		BaseClient: NewBaseClient(config),
	}
}

// Connect 连接到MCP服务器（启动进程）
func (c *StdioClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.IsConnected() {
		return ErrAlreadyConnected
	}

	// 解析环境变量
	env := os.Environ()
	if c.config.Env != nil {
		var envMap map[string]string
		if err := json.Unmarshal(c.config.Env, &envMap); err == nil {
			for k, v := range envMap {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}

	// 解析命令行参数
	var args []string
	if c.config.Args != nil {
		if err := json.Unmarshal(c.config.Args, &args); err != nil {
			return fmt.Errorf("failed to parse args: %w", err)
		}
	}

	// 创建命令
	c.cmd = exec.CommandContext(ctx, c.config.Command, args...)
	c.cmd.Env = env
	if c.config.WorkingDir != "" {
		c.cmd.Dir = c.config.WorkingDir
	}

	// 创建管道
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	c.stdout = stdout

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	c.stderr = stderr

	// 启动进程
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	// 异步读取stderr用于调试
	go c.readStderr()

	c.setConnected(true)
	return nil
}

// Disconnect 断开连接（终止进程）
func (c *StdioClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.IsConnected() {
		return nil
	}

	c.setConnected(false)

	// 关闭管道
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}

	// 终止进程
	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
		c.cmd.Wait()
	}

	return nil
}

// Initialize 初始化MCP会话
func (c *StdioClient) Initialize(ctx context.Context) (*MCPServerInfo, error) {
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
func (c *StdioClient) ListTools(ctx context.Context) ([]MCPTool, error) {
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
func (c *StdioClient) CallTool(ctx context.Context, params MCPCallToolParams) (*MCPCallToolResult, error) {
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
func (c *StdioClient) ListResources(ctx context.Context) ([]MCPResource, error) {
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
func (c *StdioClient) ReadResource(ctx context.Context, params MCPReadResourceParams) (*MCPReadResourceResult, error) {
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
func (c *StdioClient) Ping(ctx context.Context) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	var result MCPPingResult
	return c.sendRequest(ctx, "ping", MCPPingParams{}, &result)
}

// sendRequest 发送请求并接收响应
func (c *StdioClient) sendRequest(ctx context.Context, method string, params any, result any) error {
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

	// 发送请求
	c.mu.Lock()
	if _, err := c.stdin.Write(append(reqData, '\n')); err != nil {
		c.mu.Unlock()
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to write request: %w", err)
	}
	c.mu.Unlock()

	// 读取响应
	reader := bufio.NewReader(c.stdout)
	respData, err := reader.ReadBytes('\n')
	if err != nil {
		c.updateStats(false, time.Since(start))
		return fmt.Errorf("failed to read response: %w", err)
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

// readStderr 读取stderr输出（用于调试）
func (c *StdioClient) readStderr() {
	if c.stderr == nil {
		return
	}

	scanner := bufio.NewScanner(c.stderr)
	for scanner.Scan() {
		log.Printf("[MCP %s stderr] %s", c.config.Name, scanner.Text())
	}
}
