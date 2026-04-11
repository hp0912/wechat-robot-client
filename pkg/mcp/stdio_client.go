package mcp

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"wechat-robot-client/model"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// StdioClient 基于官方 go-sdk 的 Stdio 客户端封装
type StdioClient struct {
	*BaseClient
	client    *sdkmcp.Client
	session   *sdkmcp.ClientSession
	transport *sdkmcp.IOTransport
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

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if stderrPipe, err := cmd.StderrPipe(); err == nil {
		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				log.Printf("[mcp-stdio][%s] %s", c.config.Name, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				log.Printf("[mcp-stdio][%s] stderr scan error: %v", c.config.Name, err)
			}
		}()
	}

	// 追加自定义环境变量
	envList := os.Environ()

	if env, err := c.config.GetEnv(); err == nil && len(env) > 0 {
		for k, v := range env {
			envList = append(envList, k+"="+v)
		}
	}

	cmd.Env = envList

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			c.Disconnect()
			log.Printf("[mcp-stdio][%s] process exited: %v", c.config.Name, err)
		}
	}()

	filteredStdoutReader := newMCPStdoutFilter(stdoutPipe, c.config.Name)
	c.transport = &sdkmcp.IOTransport{Reader: filteredStdoutReader, Writer: stdinPipe}
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

type mcpStdoutFilter struct {
	serverName string
	raw        io.ReadCloser
	pipeReader *io.PipeReader
	pipeWriter *io.PipeWriter
}

func newMCPStdoutFilter(raw io.ReadCloser, serverName string) *mcpStdoutFilter {
	pipeReader, pipeWriter := io.Pipe()
	f := &mcpStdoutFilter{
		serverName: serverName,
		raw:        raw,
		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
	}
	go f.run()
	return f
}

func (f *mcpStdoutFilter) Read(p []byte) (int, error) {
	return f.pipeReader.Read(p)
}

func (f *mcpStdoutFilter) Close() error {
	_ = f.pipeReader.Close()
	_ = f.pipeWriter.Close()
	return f.raw.Close()
}

func (f *mcpStdoutFilter) run() {
	reader := bufio.NewReader(f.raw)
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			trimmed := strings.TrimSpace(line)
			if isLikelyJSONRPCLine(trimmed) {
				if _, werr := io.WriteString(f.pipeWriter, trimmed+"\n"); werr != nil {
					_ = f.pipeWriter.CloseWithError(werr)
					return
				}
			} else {
				log.Printf("[mcp-stdio][%s][stdout-ignored] %s", f.serverName, trimmed)
			}
		}

		if err != nil {
			if err == io.EOF {
				_ = f.pipeWriter.Close()
				return
			}
			_ = f.pipeWriter.CloseWithError(err)
			return
		}
	}
}

func isLikelyJSONRPCLine(line string) bool {
	if line == "" {
		return false
	}
	if !strings.HasPrefix(line, "{") {
		return false
	}
	return strings.Contains(line, `"jsonrpc"`) && (strings.Contains(line, `"id"`) || strings.Contains(line, `"method"`) || strings.Contains(line, `"result"`) || strings.Contains(line, `"error"`))
}
