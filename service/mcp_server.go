package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

var mcpNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type MCPServerService struct {
	ctx           context.Context
	mcpServerRepo *repository.MCPServer
}

func NewMCPServerService(ctx context.Context) *MCPServerService {
	return &MCPServerService{
		ctx:           ctx,
		mcpServerRepo: repository.NewMCPServerRepo(ctx, vars.DB),
	}
}

func (s *MCPServerService) GetMCPServers() ([]*model.MCPServer, error) {
	return s.mcpServerRepo.FindAll()
}

func (s *MCPServerService) GetMCPServer(id uint64) (*model.MCPServer, error) {
	return s.mcpServerRepo.FindByID(id)
}

func (s *MCPServerService) validateMCPServerName(mcpServer *model.MCPServer) error {
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

func (s *MCPServerService) CreateMCPServer(mcpServer *model.MCPServer) error {
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

func (s *MCPServerService) UpdateMCPServer(mcpServer *model.MCPServer) error {
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
		return vars.MCPService.ReloadServer(mcpServer.ID)
	}
	return nil
}

func (s *MCPServerService) EnableMCPServer(id uint64) error {
	if vars.MCPService == nil {
		return fmt.Errorf("MCP服务未初始化")
	}
	return vars.MCPService.EnableServer(id)
}

func (s *MCPServerService) DisableMCPServer(id uint64) error {
	if vars.MCPService == nil {
		return fmt.Errorf("MCP服务未初始化")
	}
	return vars.MCPService.DisableServer(id)
}

func (s *MCPServerService) DeleteMCPServer(mcpServer *model.MCPServer) error {
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
	if vars.MCPService != nil {
		if err := vars.MCPService.RemoveServer(mcpServer.ID); err != nil {
			fmt.Printf("停止MCP服务器时出错（将继续删除）: %v\n", err)
		}
	}

	return s.mcpServerRepo.Delete(mcpServer.ID)
}

func (s *MCPServerService) GetMCPServerTools(id uint64) ([]*sdkmcp.Tool, error) {
	if vars.MCPService == nil {
		return nil, fmt.Errorf("MCP服务未初始化")
	}
	return vars.MCPService.GetToolsByServerID(id)
}
