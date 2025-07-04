package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type SystemMessageService struct {
	ctx        context.Context
	sysmsgRepo *repository.SystemMessage
}

func NewSystemMessageService(ctx context.Context) *SystemMessageService {
	return &SystemMessageService{
		ctx:        ctx,
		sysmsgRepo: repository.NewSystemMessageRepo(ctx, vars.DB),
	}
}

func (s *SystemMessageService) GetRecentMonthMessages() ([]*model.SystemMessage, error) {
	return s.sysmsgRepo.GetRecentMonthMessages()
}
