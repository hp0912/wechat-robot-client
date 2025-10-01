package service

import (
	"context"
	"fmt"

	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type OSSSettingService struct {
	ctx             context.Context
	ossSettingsRepo *repository.OSSSettings
}

func NewOSSSettingService(ctx context.Context) *OSSSettingService {
	return &OSSSettingService{
		ctx:             ctx,
		ossSettingsRepo: repository.NewOSSSettingsRepo(ctx, vars.DB),
	}
}

func (s *OSSSettingService) GetOSSSettingService() (*model.OSSSettings, error) {
	ossSettings, err := s.ossSettingsRepo.GetOSSSettings()
	if err != nil {
		return nil, fmt.Errorf("获取OSS设置失败: %w", err)
	}
	if ossSettings == nil {
		return &model.OSSSettings{}, nil
	}
	return ossSettings, nil
}

func (s *OSSSettingService) SaveOSSSettingService(req *model.OSSSettings) error {
	if req.ID == 0 {
		ossSettings, err := s.ossSettingsRepo.GetOSSSettings()
		if err != nil {
			return err
		}
		if ossSettings != nil {
			return fmt.Errorf("OSS设置已存在，不能重复创建")
		}
		return s.ossSettingsRepo.Create(req)
	}
	return s.ossSettingsRepo.Update(req)
}
