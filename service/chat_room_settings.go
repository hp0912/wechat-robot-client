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
	return respo.GetByOwner(vars.RobotRuntime.WxID, chatRoomID)
}

func (s *ChatRoomSettingsService) GetAllEnableChatRank() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableChatRank(vars.RobotRuntime.WxID)
}

func (s *ChatRoomSettingsService) GetAllEnableAISummary() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableAISummary(vars.RobotRuntime.WxID)
}

func (s *ChatRoomSettingsService) GetAllEnableGoodMorning() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableGoodMorning(vars.RobotRuntime.WxID)
}

func (s *ChatRoomSettingsService) GetAllEnableNews() ([]*model.ChatRoomSettings, error) {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}, nil
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableNews(vars.RobotRuntime.WxID)
}

func (s *ChatRoomSettingsService) SaveChatRoomSettings(data *model.ChatRoomSettings) error {
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	if data.ID == 0 {
		data.Owner = vars.RobotRuntime.WxID
		return respo.Create(data)
	}
	return respo.Update(data)
}
