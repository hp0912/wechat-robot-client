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
	ctx     context.Context
	msgRepo *repository.Message
}

func NewChatHistoryService(ctx context.Context) *ChatHistoryService {
	return &ChatHistoryService{
		ctx:     ctx,
		msgRepo: repository.NewMessageRepo(ctx, vars.DB),
	}
}

func (s *ChatHistoryService) GetChatHistory(req dto.ChatHistoryRequest, pager appx.Pager) ([]*model.Message, int64, error) {
	return s.msgRepo.GetByContactID(req, pager)
}
