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

func (s *FriendSettingsService) GetFriendSettings(contactID string) (*model.FriendSettings, error) {
	respo := repository.NewFriendSettingsRepo(s.ctx, vars.DB)
	return respo.GetFriendSettings(contactID)
}

func (s *FriendSettingsService) SaveFriendSettings(data *model.FriendSettings) error {
	respo := repository.NewFriendSettingsRepo(s.ctx, vars.DB)
	if data.ID == 0 {
		return respo.Create(data)
	}
	return respo.Update(data)
}
