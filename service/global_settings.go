package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type GlobalSettingsService struct {
	ctx context.Context
}

func NewGlobalSettingsService(ctx context.Context) *GlobalSettingsService {
	return &GlobalSettingsService{
		ctx: ctx,
	}
}

func (s *GlobalSettingsService) GetGlobalSettings() *model.GlobalSettings {
	respo := repository.NewGlobalSettingsRepo(s.ctx, vars.DB)
	return respo.GetByOwner(vars.RobotRuntime.WxID)
}

func (s *GlobalSettingsService) SaveGlobalSettings(data *model.GlobalSettings) {
	respo := repository.NewGlobalSettingsRepo(s.ctx, vars.DB)
	respo.Update(data)
}
