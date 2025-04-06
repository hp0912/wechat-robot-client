package repository

import (
	"context"

	"gorm.io/gorm"
)

type RobotRepo struct{}

func (RobotRepo) TableName() string {
	return "robot"
}

type Robot struct {
	Base[RobotRepo]
}

func NewRobotRepo(ctx context.Context, db *gorm.DB) *Robot {
	return &Robot{
		Base[RobotRepo]{
			Ctx: ctx,
			DB:  db,
		}}
}
