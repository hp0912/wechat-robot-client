package service

import (
	"context"
	"errors"
	"fmt"
	"log"

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

func (s *OSSSettingService) UploadImageToOSS(settings *model.OSSSettings, message *model.Message) error {
	if settings.OSSProvider == "" {
		return errors.New("OSS服务商未配置")
	}
	switch settings.OSSProvider {
	case model.OSSProviderAliyun:
		if settings.AliyunOSSSettings == nil {
			return errors.New("阿里云OSS配置项未配置")
		}
		err := s.UploadImageToAliyun(settings, message)
		if err != nil {
			return err
		}
	case model.OSSProviderTencentCloud:
		if settings.TencentCloudOSSSettings == nil {
			return errors.New("腾讯云COS配置项未配置")
		}
		err := s.UploadImageToTencentCloud(settings, message)
		if err != nil {
			return err
		}
	case model.OSSProviderCloudflare:
		if settings.CloudflareR2Settings == nil {
			return errors.New("cloudflare r2配置项未配置")
		}
		err := s.UploadImageToCloudflareR2(settings, message)
		if err != nil {
			return err
		}
	default:
		log.Printf("不支持的OSS服务商: %s", settings.OSSProvider)
	}
	return nil
}

func (s *OSSSettingService) UploadImageToAliyun(settings *model.OSSSettings, message *model.Message) error {
	return nil
}

func (s *OSSSettingService) UploadImageToTencentCloud(settings *model.OSSSettings, message *model.Message) error {
	return nil
}

func (s *OSSSettingService) UploadImageToCloudflareR2(settings *model.OSSSettings, message *model.Message) error {
	return nil
}
