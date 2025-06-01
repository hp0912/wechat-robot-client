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
	return respo.GetByOwner(vars.RobotRuntime.WxID)
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

func (s *GlobalSettingsService) UpdateGlobalSettings(wxID string) error {
	respo := repository.NewGlobalSettingsRepo(s.ctx, vars.DB)
	data, err := respo.GetByOwner(wxID)
	if err != nil {
		return err
	}
	// 说明是首次登陆
	if data == nil {
		// 刚创建的机器人实例默认所有人是空
		data, err = respo.GetByOwner("")
		if err != nil {
			return err
		}
		if data == nil {
			// 说明同一个机器人实例登陆了两个微信账号
			// 随机复制一个登陆的账号的设置
			data, err = respo.GetRandomOne()
			if err != nil {
				return err
			}
			data.ID = 0 // 重置ID
			data.Owner = wxID
			err = respo.Create(data)
			if err != nil {
				return err
			}
		} else {
			data.Owner = wxID
			err = respo.Update(data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
