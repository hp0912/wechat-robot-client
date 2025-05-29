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
	// 重置公共定时任务
	vars.CronManager.Clear()
	vars.CronManager.SetGlobalSettings(data)
	vars.CronManager.Start()
}

func (s *GlobalSettingsService) UpdateGlobalSettings(wxID string) {
	respo := repository.NewGlobalSettingsRepo(s.ctx, vars.DB)
	data := respo.GetByOwner(wxID)
	// 说明是首次登陆
	if data == nil {
		// 刚创建的机器人实例默认所有人是空
		data = respo.GetByOwner("")
		if data == nil {
			// 说明同一个机器人实例登陆了两个微信账号
			// 随机复制一个登陆的账号的设置
			data = respo.GetRandomOne()
			data.ID = 0 // 重置ID
			data.Owner = wxID
			respo.Create(data)
		} else {
			data.Owner = wxID
			respo.Update(data)
		}
	}
}
