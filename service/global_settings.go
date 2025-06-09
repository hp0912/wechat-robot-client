package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type GlobalSettingsService struct {
	ctx     context.Context
	gsRespo *repository.GlobalSettings
}

func NewGlobalSettingsService(ctx context.Context) *GlobalSettingsService {
	return &GlobalSettingsService{
		ctx:     ctx,
		gsRespo: repository.NewGlobalSettingsRepo(ctx, vars.DB),
	}
}

func (s *GlobalSettingsService) GetGlobalSettings() (*model.GlobalSettings, error) {
	return s.gsRespo.GetGlobalSettings()
}

func (s *GlobalSettingsService) SaveGlobalSettings(data *model.GlobalSettings) error {
	data.FriendSyncCron = "" // 这个不允许用户修改
	err := s.gsRespo.Update(data)
	if err != nil {
		return err
	}
	// 重置公共定时任务
	newData, err := s.GetGlobalSettings()
	if err != nil {
		return err
	}
	vars.CronManager.Clear()
	vars.CronManager.SetGlobalSettings(newData)
	if vars.RobotRuntime.Status == model.RobotStatusOnline {
		vars.CronManager.Start()
	}
	return nil
}
