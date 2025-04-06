package service

import "context"

type RobotService struct {
	ctx context.Context
}

func NewDvaAppService(ctx context.Context) *RobotService {
	return &RobotService{
		ctx: ctx,
	}
}
