package mcp

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"wechat-robot-client/model"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type StreamableClient struct {
	*BaseClient
	client     *sdkmcp.Client
	session    *sdkmcp.ClientSession
	httpClient *http.Client
	url        string
	authHeader string
	headers    map[string]string
}

// NewStreamableClient 创建Streamable客户端
func NewStreamableClient(config *model.MCPServer) *StreamableClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.TLSSkipVerify != nil && *config.TLSSkipVerify,
			},
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: false,
		},
	}

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
		httpClient: httpClient,
		url:        config.URL,
		authHeader: authHeader,
		headers:    headers,
	}
}

// ensureAuth 基于 MCP 授权规范进行最小实现：
// - 遵循 RFC9728: 解析 401 的 WWW-Authenticate，定位资源元数据 URL
// - 拉取资源元数据，定位 authorization_servers
// - 尝试通过客户端凭据/静态 Token（配置）构造 Bearer 令牌
// 说明：出于简化，本实现不内置完整 OAuth 授权码流程，仅支持：
// 1) 使用配置中的静态 Bearer Token；
// 2) 若无静态 Token，则优雅返回，由上层引导用户完成授权。
func (c *StreamableClient) ensureAuth(ctx context.Context) error {
	// 若已有 header，则直接返回
	if c.authHeader != "" {
		return nil
	}
	// 1) 首选配置中的 Token
	if c.config.AuthType == model.MCPAuthTypeBearer && c.config.AuthToken != "" {
		c.authHeader = "Bearer " + c.config.AuthToken
		return nil
	}

	// 2) 按规范进行一次探测：请求服务器根或已知端点，期待 401 + WWW-Authenticate
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}
	// 附加自定义头（若有）
	for k, v := range c.headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 期望 401 返回
	if resp.StatusCode != http.StatusUnauthorized {
		// 尝试解析 body 是否包含受保护资源元数据（少数服务直接暴露）
		body, _ := io.ReadAll(resp.Body)
		var meta map[string]any
		if json.Unmarshal(body, &meta) == nil {
			if setAuthFromProtectedResource(meta, c) {
				return nil
			}
		}
		return errors.New("authorization discovery not available (no 401)")
	}
	// 解析 WWW-Authenticate，查找资源元数据 URL（RFC9728 S5.1）
	www := resp.Header.Get("WWW-Authenticate")
	metaURL := parseProtectedResourceMetadataURL(www)
	if metaURL == "" {
		return errors.New("missing protected resource metadata URL in WWW-Authenticate")
	}
	// 拉取受保护资源元数据，定位 authorization_servers
	mreq, err := http.NewRequestWithContext(ctx, http.MethodGet, metaURL, nil)
	if err != nil {
		return err
	}
	mresp, err := c.httpClient.Do(mreq)
	if err != nil {
		return err
	}
	defer mresp.Body.Close()
	if mresp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch protected resource metadata: %s", mresp.Status)
	}
	body, _ := io.ReadAll(mresp.Body)
	var meta map[string]any
	if err := json.Unmarshal(body, &meta); err != nil {
		return err
	}
	if !setAuthFromProtectedResource(meta, c) {
		// 没有可用的自动令牌路径，维持无 Token 状态，由上层处理交互授权
		return errors.New("no usable authorization server or token in metadata")
	}
	return nil
}

// parseProtectedResourceMetadataURL 尝试从 WWW-Authenticate 中提取 resource metadata URL
func parseProtectedResourceMetadataURL(header string) string {
	if header == "" {
		return ""
	}
	// 常见形式: Bearer authorization_uri="https://example.com/.well-known/oauth-protected-resource"
	// 或 resource_metadata="..."
	// 简单解析 k="v" 对
	parts := strings.Split(header, ",")
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(kv[0]))
		val := strings.Trim(strings.TrimSpace(kv[1]), "\"")
		if key == "authorization_uri" || key == "resource_metadata" {
			// 校验 URL 格式
			if u, err := url.Parse(val); err == nil && u.Scheme != "" && u.Host != "" {
				return val
			}
		}
	}
	return ""
}

// setAuthFromProtectedResource 最小实现：若 metadata 中含 authorization_servers，则仅记录信息；
// 若本地已有静态 Token 则直接设置；完整 OAuth 授权流程交由上层配置/UI。
func setAuthFromProtectedResource(meta map[string]any, c *StreamableClient) bool {
	if c.config.AuthToken != "" {
		c.authHeader = "Bearer " + c.config.AuthToken
		return true
	}
	// 记录 authorization_servers 等信息以备后用（当前不实现动态注册/授权码流）
	// 这里只做存在性检查
	if v, ok := meta["authorization_servers"]; ok && v != nil {
		return true
	}
	return false
}

// Connect 连接到MCP服务器
func (c *StreamableClient) Connect(ctx context.Context) error {
	if c.IsConnected() {
		return ErrAlreadyConnected
	}

	// 若需要鉴权但未携带令牌，则尝试基于规范进行发现并获取访问令牌
	if c.config.AuthType == model.MCPAuthTypeBearer && c.authHeader == "" {
		_ = c.ensureAuth(ctx) // 最多一次最佳努力，不阻断老服务
	}

	transport := &sdkmcp.StreamableClientTransport{Endpoint: c.url}
	// 注入 HTTP 客户端与请求头，兼容不同 SDK 版本字段
	hdr := make(http.Header)
	if c.authHeader != "" {
		hdr.Set("Authorization", c.authHeader)
	}
	for k, v := range c.headers {
		if v != "" {
			hdr.Set(k, v)
		}
	}
	tv := reflect.ValueOf(transport).Elem()
	if f := tv.FieldByName("HTTPClient"); f.IsValid() && f.CanSet() && f.Type().String() == "*http.Client" {
		f.Set(reflect.ValueOf(c.httpClient))
	}
	for _, name := range []string{"Headers", "Header", "RequestHeaders", "RequestHeader"} {
		f := tv.FieldByName(name)
		if !f.IsValid() || !f.CanSet() {
			continue
		}
		switch f.Type().String() {
		case "http.Header":
			f.Set(reflect.ValueOf(hdr))
		case "map[string][]string":
			m := map[string][]string(hdr)
			f.Set(reflect.ValueOf(m))
		case "map[string]string":
			ms := make(map[string]string, len(hdr))
			for k, vals := range hdr {
				if len(vals) > 0 {
					ms[k] = vals[0]
				}
			}
			f.Set(reflect.ValueOf(ms))
		}
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
		// 连接失败时，若可能为 401，尝试一次鉴权发现与重试
		if c.config.AuthType == model.MCPAuthTypeBearer && c.authHeader == "" {
			if derr := c.ensureAuth(ctx); derr == nil && c.authHeader != "" {
				// 重试一次
				sess, err = c.client.Connect(ctx, transport, nil)
				if err == nil {
					c.session = sess
					c.setConnected(true)
					return nil
				}
			}
		}
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
