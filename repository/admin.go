package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type RobotAdmin struct {
	Base[model.RobotAdmin]
}

func NewRobotAdminRepo(ctx context.Context, db *gorm.DB) *RobotAdmin {
	return &RobotAdmin{
		Base[model.RobotAdmin]{
			Ctx: ctx,
			DB:  db,
		}}
}
