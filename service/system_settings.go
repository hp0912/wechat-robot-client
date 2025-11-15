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
	var err error
	if req.ID == 0 {
		systemSettings, err := s.systemSettingsRepo.GetSystemSettings()
		if err != nil {
			return err
		}
		if systemSettings != nil {
			return fmt.Errorf("系统设置已存在，不能重复创建")
		}
		err = s.systemSettingsRepo.Create(req)
		if err != nil {
			return fmt.Errorf("创建系统设置失败: %w", err)
		}
	} else {
		err = s.systemSettingsRepo.Update(req)
		if err != nil {
			return fmt.Errorf("更新系统设置失败: %w", err)
		}
	}

	// 更新全局 webhook 配置
	s.updateWebhookConfig(req)
	return nil
}

func (s *SystemSettingService) updateWebhookConfig(req *model.SystemSettings) {
	if req.WebhookURL != nil {
		vars.Webhook.URL = *req.WebhookURL
	} else {
		vars.Webhook.URL = ""
	}
	if req.WebhookHeaders != nil {
		vars.Webhook.Headers = req.WebhookHeaders
	} else {
		vars.Webhook.Headers = nil
	}
}
