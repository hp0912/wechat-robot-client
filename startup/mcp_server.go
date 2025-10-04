package startup

import (
	"context"

	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

func InitMCPService() error {
	ctx := context.Background()

	messageService := service.NewMessageService(ctx)
	messageSender := service.NewMessageSenderAdapter(messageService)
	vars.MCPService = service.NewMCPService(ctx, vars.DB, messageSender)

	err := vars.MCPService.Initialize()
	if err != nil {
		return err
	}
	return nil
}
