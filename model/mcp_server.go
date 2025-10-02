package model

import (
	"time"

	"gorm.io/datatypes"
)

// MCPTransportType MCP服务器传输类型
type MCPTransportType string

const (
	MCPTransportTypeStdio MCPTransportType = "stdio" // 命令行模式（标准输入输出）
	MCPTransportTypeSSE   MCPTransportType = "sse"   // Server-Sent Events模式
	MCPTransportTypeHTTP  MCPTransportType = "http"  // HTTP模式
	MCPTransportTypeWS    MCPTransportType = "ws"    // WebSocket模式
)

// MCPAuthType MCP认证类型
type MCPAuthType string

const (
	MCPAuthTypeNone   MCPAuthType = "none"   // 无认证
	MCPAuthTypeBearer MCPAuthType = "bearer" // Bearer Token认证
	MCPAuthTypeBasic  MCPAuthType = "basic"  // Basic认证
	MCPAuthTypeAPIKey MCPAuthType = "apikey" // API Key认证
)

// MCPServer MCP服务器配置表
type MCPServer struct {
	ID          uint64           `gorm:"column:id;primaryKey;autoIncrement;comment:MCP服务器配置表主键ID" json:"id"`
	Name        string           `gorm:"column:name;type:varchar(100);not null;comment:MCP服务器名称" json:"name"`
	Description string           `gorm:"column:description;type:varchar(500);default:'';comment:MCP服务器描述" json:"description"`
	Transport   MCPTransportType `gorm:"column:transport;type:enum('stdio','sse','http','ws');not null;comment:传输类型：stdio-命令行，sse-SSE，http-HTTP，ws-WebSocket" json:"transport"`
	Enabled     *bool            `gorm:"column:enabled;default:true;comment:是否启用该MCP服务器" json:"enabled"`
	Priority    int              `gorm:"column:priority;default:0;comment:优先级，数字越大优先级越高" json:"priority"`

	// Stdio模式配置（命令行模式）
	Command    string         `gorm:"column:command;type:varchar(255);default:'';comment:命令行模式的可执行命令" json:"command"`
	Args       datatypes.JSON `gorm:"column:args;type:json;comment:命令行参数数组" json:"args"` // []string
	WorkingDir string         `gorm:"column:working_dir;type:varchar(500);default:'';comment:工作目录" json:"working_dir"`
	Env        datatypes.JSON `gorm:"column:env;type:json;comment:环境变量键值对" json:"env"` // map[string]string

	// 网络模式配置（SSE/HTTP/WS共用）
	URL           string         `gorm:"column:url;type:varchar(500);default:'';comment:服务器URL地址（SSE/HTTP/WS模式）" json:"url"`
	AuthType      MCPAuthType    `gorm:"column:auth_type;type:enum('none','bearer','basic','apikey');default:'none';comment:认证类型" json:"auth_type"`
	AuthToken     string         `gorm:"column:auth_token;type:varchar(500);default:'';comment:认证令牌（Bearer Token或API Key）" json:"auth_token"`
	AuthUsername  string         `gorm:"column:auth_username;type:varchar(100);default:'';comment:Basic认证用户名" json:"auth_username"`
	AuthPassword  string         `gorm:"column:auth_password;type:varchar(255);default:'';comment:Basic认证密码" json:"auth_password"`
	Headers       datatypes.JSON `gorm:"column:headers;type:json;comment:自定义HTTP请求头" json:"headers"` // map[string]string
	TLSSkipVerify *bool          `gorm:"column:tls_skip_verify;default:false;comment:是否跳过TLS证书验证" json:"tls_skip_verify"`

	// 超时和重连配置
	ConnectTimeout    int   `gorm:"column:connect_timeout;default:30;comment:连接超时时间（秒）" json:"connect_timeout"`
	ReadTimeout       int   `gorm:"column:read_timeout;default:60;comment:读取超时时间（秒）" json:"read_timeout"`
	WriteTimeout      int   `gorm:"column:write_timeout;default:60;comment:写入超时时间（秒）" json:"write_timeout"`
	MaxRetries        int   `gorm:"column:max_retries;default:3;comment:最大重试次数" json:"max_retries"`
	RetryInterval     int   `gorm:"column:retry_interval;default:5;comment:重试间隔时间（秒）" json:"retry_interval"`
	HeartbeatEnable   *bool `gorm:"column:heartbeat_enable;default:true;comment:是否启用心跳检测" json:"heartbeat_enable"`
	HeartbeatInterval int   `gorm:"column:heartbeat_interval;default:30;comment:心跳间隔时间（秒）" json:"heartbeat_interval"`

	// 高级配置
	Capabilities datatypes.JSON `gorm:"column:capabilities;type:json;comment:MCP服务器能力配置" json:"capabilities"` // 服务器支持的能力
	CustomConfig datatypes.JSON `gorm:"column:custom_config;type:json;comment:自定义配置项" json:"custom_config"`   // 其他自定义配置
	Tags         datatypes.JSON `gorm:"column:tags;type:json;comment:标签列表" json:"tags"`                       // []string，用于分类和过滤

	// 状态追踪
	LastConnectedAt *time.Time `gorm:"column:last_connected_at;type:datetime;comment:最后连接成功时间" json:"last_connected_at"`
	LastError       string     `gorm:"column:last_error;type:text;comment:最后一次错误信息" json:"last_error"`
	ConnectionCount int64      `gorm:"column:connection_count;default:0;comment:累计连接次数" json:"connection_count"`
	ErrorCount      int64      `gorm:"column:error_count;default:0;comment:累计错误次数" json:"error_count"`

	// 时间戳
	CreatedAt time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime;index;comment:软删除时间" json:"deleted_at,omitempty"`
}

// TableName 设置表名
func (MCPServer) TableName() string {
	return "mcp_servers"
}

// IsStdio 判断是否为命令行模式
func (m *MCPServer) IsStdio() bool {
	return m.Transport == MCPTransportTypeStdio
}

// IsNetworkMode 判断是否为网络模式（SSE/HTTP/WS）
func (m *MCPServer) IsNetworkMode() bool {
	return m.Transport == MCPTransportTypeSSE ||
		m.Transport == MCPTransportTypeHTTP ||
		m.Transport == MCPTransportTypeWS
}

// NeedsAuth 判断是否需要认证
func (m *MCPServer) NeedsAuth() bool {
	return m.AuthType != MCPAuthTypeNone && m.AuthType != ""
}
