package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/mcp"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

var mcpNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

type MCPService struct {
	ctx           context.Context
	manager       *mcp.MCPManager
	mcpServerRepo *repository.MCPServer
}

func NewMCPService(ctx context.Context) *MCPService {
	manager := mcp.NewMCPManager(vars.DB)
	return &MCPService{
		ctx:           ctx,
		manager:       manager,
		mcpServerRepo: repository.NewMCPServerRepo(ctx, vars.DB),
	}
}

func (s *MCPService) GetMCPServers() ([]*model.MCPServer, error) {
	return s.mcpServerRepo.FindAll()
}

func (s *MCPService) GetMCPServer(id uint64) (*model.MCPServer, error) {
	return s.mcpServerRepo.FindByID(id)
}

func (s *MCPService) validateMCPServerName(mcpServer *model.MCPServer) error {
	if mcpServer == nil {
		return fmt.Errorf("MCP服务器对象不能为空")
	}
	name := strings.TrimSpace(mcpServer.Name)
	if name == "" {
		return fmt.Errorf("MCP服务器名称不能为空")
	}
	if !mcpNameRe.MatchString(name) {
		return fmt.Errorf("无效的MCP服务器名称：%q。只允许字母、数字、下划线，且必须以字母或下划线开头", name)
	}
	mcpServer.Name = name
	return nil
}

func (s *MCPService) CreateMCPServer(mcpServer *model.MCPServer) error {
	if err := s.validateMCPServerName(mcpServer); err != nil {
		return err
	}
	if mcpServer.HeartbeatEnable != nil && *mcpServer.HeartbeatEnable && mcpServer.HeartbeatInterval < 60 {
		return fmt.Errorf("心跳间隔不能小于60秒")
	}
	now := time.Now()
	mcpServer.CreatedAt = &now
	mcpServer.UpdatedAt = &now
	return s.mcpServerRepo.Create(mcpServer)
}

func (s *MCPService) UpdateMCPServer(mcpServer *model.MCPServer) error {
	if err := s.validateMCPServerName(mcpServer); err != nil {
		return err
	}
	if mcpServer.HeartbeatEnable != nil && *mcpServer.HeartbeatEnable && mcpServer.HeartbeatInterval < 60 {
		return fmt.Errorf("心跳间隔不能小于60秒")
	}
	server, err := s.mcpServerRepo.FindByID(mcpServer.ID)
	if err != nil {
		return err
	}
	if server == nil {
		return fmt.Errorf("MCP服务器不存在")
	}
	if server.IsBuiltIn != nil && *server.IsBuiltIn {
		return fmt.Errorf("官方 MCP 服务不支持编辑")
	}
	if server.Transport != mcpServer.Transport {
		return fmt.Errorf("不允许修改MCP服务器类型")
	}

	now := time.Now()
	mcpServer.UpdatedAt = &now
	err = s.mcpServerRepo.Update(mcpServer)
	if err != nil {
		return err
	}

	server, err = s.mcpServerRepo.FindByID(mcpServer.ID)
	if err != nil {
		return err
	}
	if server.Enabled != nil && *server.Enabled {
		return s.ReloadServer(mcpServer.ID)
	}
	return nil
}

func (s *MCPService) EnableMCPServer(id uint64) error {
	if vars.Agent == nil {
		return fmt.Errorf("MCP服务未初始化")
	}
	return s.EnableServer(id)
}

func (s *MCPService) DisableMCPServer(id uint64) error {
	if vars.Agent == nil {
		return fmt.Errorf("MCP服务未初始化")
	}
	return s.DisableServer(id)
}

func (s *MCPService) DeleteMCPServer(mcpServer *model.MCPServer) error {
	if mcpServer == nil || mcpServer.ID == 0 {
		return fmt.Errorf("参数异常")
	}
	currentMCPServer, err := s.mcpServerRepo.FindByID(mcpServer.ID)
	if err != nil {
		return fmt.Errorf("查询MCP服务器失败: %w", err)
	}
	if currentMCPServer == nil {
		return fmt.Errorf("MCP服务器不存在")
	}
	if currentMCPServer.IsBuiltIn != nil && *currentMCPServer.IsBuiltIn {
		return fmt.Errorf("内置MCP服务器不允许删除")
	}
	if vars.Agent != nil {
		if err := s.RemoveServer(mcpServer.ID); err != nil {
			fmt.Printf("停止MCP服务器时出错（将继续删除）: %v\n", err)
		}
	}

	return s.mcpServerRepo.Delete(mcpServer.ID)
}

// RemoveServer 移除MCP服务器
func (s *MCPService) RemoveServer(serverID uint64) error {
	// 从管理器移除
	s.manager.RemoveServer(serverID)

	// 从数据库删除
	return s.mcpServerRepo.Delete(serverID)
}

// UpdateServer 更新MCP服务器配置
func (s *MCPService) UpdateServer(server *model.MCPServer) error {
	// 更新数据库
	if err := s.mcpServerRepo.Update(server); err != nil {
		return fmt.Errorf("failed to update server in database: %w", err)
	}

	// 重新加载
	if server.Enabled != nil && *server.Enabled {
		if err := s.manager.ReloadServer(server.ID); err != nil {
			return fmt.Errorf("failed to reload server: %w", err)
		}
	} else {
		// 如果禁用，则断开连接
		s.manager.RemoveServer(server.ID)
	}

	return nil
}

// EnableServer 启用服务器
func (s *MCPService) EnableServer(serverID uint64) error {
	return s.manager.EnableServer(serverID)
}

// DisableServer 禁用服务器
func (s *MCPService) DisableServer(serverID uint64) error {
	return s.manager.DisableServer(serverID)
}

// GetServerByID 获取服务器配置
func (s *MCPService) GetServerByID(serverID uint64) (*model.MCPServer, error) {
	return s.mcpServerRepo.FindByID(serverID)
}

// GetAllServers 获取所有服务器配置
func (s *MCPService) GetAllServers() ([]*model.MCPServer, error) {
	return s.mcpServerRepo.FindAll()
}

// GetEnabledServers 获取已启用的服务器
func (s *MCPService) GetEnabledServers() ([]*model.MCPServer, error) {
	return s.mcpServerRepo.FindEnabled()
}

// GetServerStats 获取服务器统计信息
func (s *MCPService) GetServerStats(serverID uint64) (*mcp.MCPConnectionStats, error) {
	return s.manager.GetServerStats(serverID)
}

// GetActiveServerCount 获取活动服务器数量
func (s *MCPService) GetActiveServerCount() int {
	return s.manager.GetActiveServerCount()
}

// ReloadServer 重新加载服务器
func (s *MCPService) ReloadServer(serverID uint64) error {
	return s.manager.ReloadServer(serverID)
}

func (s *MCPService) GetMCPServerTools(id uint64) ([]*sdkmcp.Tool, error) {
	if vars.Agent == nil {
		return nil, fmt.Errorf("MCP服务未初始化")
	}
	return s.GetToolsByServerID(id)
}

// GetToolsByServerID 获取指定MCP服务器提供的工具列表（MCP SDK原始格式）
func (s *MCPService) GetToolsByServerID(serverID uint64) ([]*sdkmcp.Tool, error) {
	// 检查服务器是否存在且已启用
	server, err := s.mcpServerRepo.FindByID(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to find server: %w", err)
	}
	if server == nil {
		return nil, fmt.Errorf("server not found")
	}
	if server.Enabled == nil || !*server.Enabled {
		return nil, fmt.Errorf("server is not enabled")
	}

	// 通过manager获取客户端并列出工具
	client, err := s.manager.GetClient(serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	tools, err := client.ListTools(s.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return tools, nil
}
