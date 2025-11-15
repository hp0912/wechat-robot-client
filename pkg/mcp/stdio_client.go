package mcp

import (
	"context"
	"os/exec"
	"time"

	"wechat-robot-client/model"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StdioClient 基于官方 go-sdk 的 Stdio 客户端封装
type StdioClient struct {
	*BaseClient
	client    *sdkmcp.Client
	session   *sdkmcp.ClientSession
	transport *sdkmcp.CommandTransport
}

func NewStdioClient(config *model.MCPServer) *StdioClient {
	return &StdioClient{BaseClient: NewBaseClient(config)}
}

func (c *StdioClient) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return ErrAlreadyConnected
	}

	args, _ := c.config.GetArgs()
	cmd := exec.CommandContext(ctx, c.config.Command, args...)
	if c.config.WorkingDir != "" {
		cmd.Dir = c.config.WorkingDir
	}
	if env, err := c.config.GetEnv(); err == nil && len(env) > 0 {
		// 追加自定义环境变量
		envList := cmd.Env
		for k, v := range env {
			envList = append(envList, k+"="+v)
		}
		cmd.Env = envList
	}

	c.transport = &sdkmcp.CommandTransport{Command: cmd}
	c.client = sdkmcp.NewClient(&sdkmcp.Implementation{Name: "wechat-robot-mcp-client", Version: "1.0.0"}, nil)

	sess, err := c.client.Connect(ctx, c.transport, nil)
	if err != nil {
		return err
	}
	c.session = sess
	c.setConnected(true)
	return nil
}

func (c *StdioClient) Disconnect() error {
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

func (c *StdioClient) Initialize(ctx context.Context) (*MCPServerInfo, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}
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

func (c *StdioClient) ListTools(ctx context.Context) ([]*sdkmcp.Tool, error) {
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

func (c *StdioClient) CallTool(ctx context.Context, params *sdkmcp.CallToolParams) (*sdkmcp.CallToolResult, error) {
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

func (c *StdioClient) ListResources(ctx context.Context) ([]*sdkmcp.Resource, error) {
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

func (c *StdioClient) ReadResource(ctx context.Context, params *sdkmcp.ReadResourceParams) (*sdkmcp.ReadResourceResult, error) {
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

func (c *StdioClient) Ping(ctx context.Context) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	start := time.Now()
	err := c.session.Ping(ctx, &sdkmcp.PingParams{})
	c.updateStats(err == nil, time.Since(start))

	return err
}
