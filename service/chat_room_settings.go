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

func (s *ChatRoomSettingsService) SaveChatRoomSettings(data *model.ChatRoomSettings) {
	respo := repository.NewChatRoomSettingsRepo(s.ctx, vars.DB)
	if data.ID == 0 {
		data.Owner = vars.RobotRuntime.WxID
		respo.Create(data)
	} else {
		respo.Update(data)
	}
}
