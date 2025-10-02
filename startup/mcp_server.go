package startup

import (
	"context"

	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

func InitMCPService() (*service.MCPService, error) {
	ctx := context.Background()
	mcpService := service.NewMCPService(ctx, vars.DB)
	err := mcpService.Initialize()
	if err != nil {
		return nil, err
	}
	return mcpService, nil
}
