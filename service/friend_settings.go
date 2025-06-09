package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type FriendSettingsService struct {
	ctx     context.Context
	fsRespo *repository.FriendSettings
}

func NewFriendSettingsService(ctx context.Context) *FriendSettingsService {
	return &FriendSettingsService{
		ctx:     ctx,
		fsRespo: repository.NewFriendSettingsRepo(ctx, vars.DB),
	}
}

func (s *FriendSettingsService) GetFriendSettings(contactID string) (*model.FriendSettings, error) {
	return s.fsRespo.GetFriendSettings(contactID)
}

func (s *FriendSettingsService) SaveFriendSettings(data *model.FriendSettings) error {
	if data.ID == 0 {
		return s.fsRespo.Create(data)
	}
	return s.fsRespo.Update(data)
}
