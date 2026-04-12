package startup

import (
	"context"

	"wechat-robot-client/service"
	"wechat-robot-client/vars"
)

func InitAgent() error {
	ctx := context.Background()
	vars.Agent = service.NewAgentService(ctx, vars.DB, vars.KnowledgeService)
	err := vars.Agent.Initialize()
	if err != nil {
		return err
	}
	return nil
}
