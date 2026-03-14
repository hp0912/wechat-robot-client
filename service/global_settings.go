package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type GlobalSettingsService struct {
	ctx    context.Context
	gsRepo *repository.GlobalSettings
}

func NewGlobalSettingsService(ctx context.Context) *GlobalSettingsService {
	return &GlobalSettingsService{
		ctx:    ctx,
		gsRepo: repository.NewGlobalSettingsRepo(ctx, vars.DB),
	}
}

func (s *GlobalSettingsService) GetGlobalSettings() (*model.GlobalSettings, error) {
	return s.gsRepo.GetGlobalSettings()
}

func (s *GlobalSettingsService) SaveGlobalSettings(data *model.GlobalSettings) error {
	data.FriendSyncCron = "" // 这个不允许用户修改
	err := s.gsRepo.Update(data)
	if err != nil {
		return err
	}
	// 重新读取最新的完整配置，通知所有观察者
	newData, err := s.GetGlobalSettings()
	if err != nil {
		return err
	}
	vars.SettingsObserver.NotifyAll(newData)
	return nil
}
