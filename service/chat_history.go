package service

import (
	"context"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type ChatHistoryService struct {
	ctx context.Context
}

func NewChatHistoryService(ctx context.Context) *ChatHistoryService {
	return &ChatHistoryService{
		ctx: ctx,
	}
}

func (s *ChatHistoryService) GetChatHistory(req dto.ChatHistoryRequest, pager appx.Pager) ([]*model.Message, int64, error) {
	req.Owner = vars.RobotRuntime.WxID
	respo := repository.NewMessageRepo(s.ctx, vars.DB)
	return respo.GetByContactID(req, pager)
}
