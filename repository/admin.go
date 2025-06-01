package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type RobotAdmin struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewRobotAdminRepo(ctx context.Context, db *gorm.DB) *RobotAdmin {
	return &RobotAdmin{
		Ctx: ctx,
		DB:  db,
	}
}

func (r *RobotAdmin) GetByRobotID(robotID int64) (*model.RobotAdmin, error) {
	var robotAdmin model.RobotAdmin
	err := r.DB.WithContext(r.Ctx).Where("id = ?", robotID).First(&robotAdmin).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &robotAdmin, nil
}

func (r *RobotAdmin) Update(robot *model.RobotAdmin) error {
	return r.DB.WithContext(r.Ctx).Updates(robot).Error
}
