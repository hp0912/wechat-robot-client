package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type FriendSettingsService struct {
	ctx context.Context
}

func NewFriendSettingsService(ctx context.Context) *FriendSettingsService {
	return &FriendSettingsService{
		ctx: ctx,
	}
}

func (s *FriendSettingsService) GetFriendSettings(contactID string) *model.FriendSettings {
	respo := repository.NewFriendSettingsRepo(s.ctx, vars.DB)
	return respo.GetByOwner(vars.RobotRuntime.WxID, contactID)
}

func (s *FriendSettingsService) SaveFriendSettings(data *model.FriendSettings) {
	respo := repository.NewFriendSettingsRepo(s.ctx, vars.DB)
	if data.ID == 0 {
		data.Owner = vars.RobotRuntime.WxID
		respo.Create(data)
	} else {
		respo.Update(data)
	}
}
