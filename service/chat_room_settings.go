package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type ChatRoomSettingsService struct {
	ctx context.Context
}

func NewChatRoomSettingsService(ctx context.Context) *ChatRoomSettingsService {
	return &ChatRoomSettingsService{
		ctx: ctx,
	}
}

func (s *ChatRoomSettingsService) GetChatRoomSettings(chatRoomID string) (*model.ChatRoomSettings, error) {
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetChatRoomSettings(chatRoomID)
}

func (s *ChatRoomSettingsService) GetAllEnableChatRank() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableChatRank()
}

func (s *ChatRoomSettingsService) GetAllEnableAISummary() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableAISummary()
}

func (s *ChatRoomSettingsService) GetAllEnableGoodMorning() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableGoodMorning()
}

func (s *ChatRoomSettingsService) GetAllEnableNews() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableNews()
}

func (s *ChatRoomSettingsService) SaveChatRoomSettings(data *model.ChatRoomSettings) error {
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	if data.ID == 0 {
		return respo.Create(data)
	}
	return respo.Update(data)
}
