package service

import (
	"context"
	"wechat-robot-client/model"
	"wechat-robot-client/repository"
	"wechat-robot-client/vars"
)

type AdminService struct {
	ctx            context.Context
	robotAdminRepo *repository.RobotAdmin
}

func NewAdminService(ctx context.Context) *AdminService {
	return &AdminService{
		ctx:            ctx,
		robotAdminRepo: repository.NewRobotAdminRepo(ctx, vars.AdminDB),
	}
}

func (s *AdminService) GetRobotByID(robotID int64) (*model.RobotAdmin, error) {
	return s.robotAdminRepo.GetByRobotID(robotID)
}
