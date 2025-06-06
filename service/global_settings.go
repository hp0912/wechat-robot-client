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

func (s *GlobalSettingsService) GetGlobalSettings() (*model.GlobalSettings, error) {
	respo := repository.NewGlobalSettingsRepo(s.ctx, vars.DB)
	return respo.GetGlobalSettings()
}

func (s *GlobalSettingsService) SaveGlobalSettings(data *model.GlobalSettings) error {
	respo := repository.NewGlobalSettingsRepo(s.ctx, vars.DB)
	data.FriendSyncCron = "" // 这个不允许用户修改
	err := respo.Update(data)
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
