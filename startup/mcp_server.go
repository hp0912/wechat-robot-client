package startup

import (
	"context"

	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

func InitMCPService() error {
	ctx := context.Background()
	vars.MCPService = service.NewMCPService(ctx, vars.DB)
	err := vars.MCPService.Initialize()
	if err != nil {
		return err
	}
	return nil
}
