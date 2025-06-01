package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type AdminService struct {
	ctx context.Context
}

func NewAdminService(ctx context.Context) *AdminService {
	return &AdminService{
		ctx: ctx,
	}
}

func (s *AdminService) GetRobotByID(robotID int64) (*model.RobotAdmin, error) {
	robotRespo := repository.NewRobotAdminRepo(s.ctx, vars.AdminDB)
	return robotRespo.GetByRobotID(robotID)
}
