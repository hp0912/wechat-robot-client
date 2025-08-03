package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type MomentComment struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewMomentCommentRepo(ctx context.Context, db *gorm.DB) *MomentComment {
	return &MomentComment{
		Ctx: ctx,
		DB:  db,
	}
}

func (respo *MomentComment) Create(data *model.MomentComment) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *MomentComment) Update(data *model.MomentComment) error {
	return respo.DB.WithContext(respo.Ctx).Updates(data).Error
}
