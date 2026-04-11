package service

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"log"
// 	"strings"

// 	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
// 	"github.com/sashabaranov/go-openai"
// 	"gorm.io/gorm"

// 	"wechat-robot-client/interface/ai"
// 	"wechat-robot-client/model"
// 	"wechat-robot-client/pkg/mcp"
// 	"wechat-robot-client/pkg/robotctx"
// 	"wechat-robot-client/pkg/skills"
// 	"wechat-robot-client/repository"
// 	"wechat-robot-client/vars"
// )

// type MCPService struct {
// 	ctx           context.Context
// 	db            *gorm.DB
// 	manager       *mcp.MCPManager
// 	mcpServerRepo *repository.MCPServer
// }

// func NewMCPService(ctx context.Context, db *gorm.DB) *MCPService {
// 	manager := mcp.NewMCPManager(db)
// 	repo := repository.NewMCPServerRepo(ctx, db)

// 	return &MCPService{
// 		ctx:           ctx,
// 		db:            db,
// 		manager:       manager,
// 		mcpServerRepo: repo,
// 	}
// }

// // GetAllTools 获取所有可用工具（OpenAI格式）
// func (s *MCPService) GetAllTools() ([]openai.Tool, error) {
// 	return s.manager.GetOpenAITools(s.ctx)
// }

// // AddServer 添加MCP服务器
// func (s *MCPService) AddServer(server *model.MCPServer) error {
// 	// 保存到数据库
// 	if err := s.mcpServerRepo.Create(server); err != nil {
// 		return fmt.Errorf("failed to create server in database: %w", err)
// 	}

// 	// 如果启用，则连接
// 	if server.Enabled != nil && *server.Enabled {
// 		if err := s.manager.AddServer(server); err != nil {
// 			return fmt.Errorf("failed to connect server: %w", err)
// 		}
// 	}

// 	return nil
// }
