package service

import (
	"context"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type SystemSettingService struct {
	ctx                context.Context
	systemSettingsRepo *repository.SystemSettings
}

func NewSystemSettingService(ctx context.Context) *SystemSettingService {
	return &SystemSettingService{
		ctx:                ctx,
		systemSettingsRepo: repository.NewSystemSettingsRepo(ctx, vars.DB),
	}
}

func (s *SystemSettingService) GetSystemSettings() (*model.SystemSettings, error) {
	return nil, nil
}

func (s *SystemSettingService) SaveSystemSettings(req *model.SystemSettings) error {
	return nil
}
