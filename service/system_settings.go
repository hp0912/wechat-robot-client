package service

import (
	"context"
	"fmt"

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
	systemSettings, err := s.systemSettingsRepo.GetSystemSettings()
	if err != nil {
		return nil, fmt.Errorf("获取系统设置失败: %w", err)
	}
	if systemSettings == nil {
		return &model.SystemSettings{}, nil
	}
	return systemSettings, nil
}

func (s *SystemSettingService) SaveSystemSettings(req *model.SystemSettings) error {
	if req.ID == 0 {
		systemSettings, err := s.systemSettingsRepo.GetSystemSettings()
		if err != nil {
			return err
		}
		if systemSettings != nil {
			return fmt.Errorf("系统设置已存在，不能重复创建")
		}
		return s.systemSettingsRepo.Create(req)
	}
	return s.systemSettingsRepo.Update(req)
}
