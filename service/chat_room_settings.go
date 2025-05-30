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

func (s *ChatRoomSettingsService) GetChatRoomSettings(chatRoomID string) *model.ChatRoomSettings {
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetByOwner(vars.RobotRuntime.WxID, chatRoomID)
}

func (s *ChatRoomSettingsService) GetAllEnableChatRank() []*model.ChatRoomSettings {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableChatRank(vars.RobotRuntime.WxID)
}

func (s *ChatRoomSettingsService) GetAllEnableGoodMorning() []*model.ChatRoomSettings {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableGoodMorning(vars.RobotRuntime.WxID)
}

func (s *ChatRoomSettingsService) GetAllEnableNews() []*model.ChatRoomSettings {
	if vars.RobotRuntime.Status == model.RobotStatusOffline {
		return []*model.ChatRoomSettings{}
	}
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	return respo.GetAllEnableNews(vars.RobotRuntime.WxID)
}

func (s *ChatRoomSettingsService) SaveChatRoomSettings(data *model.ChatRoomSettings) {
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	if data.ID == 0 {
		data.Owner = vars.RobotRuntime.WxID
		respo.Create(data)
	} else {
		respo.Update(data)
	}
}
