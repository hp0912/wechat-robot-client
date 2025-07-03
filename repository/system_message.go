package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type SystemMessage struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewSystemMessageRepo(ctx context.Context, db *gorm.DB) *SystemMessage {
	return &SystemMessage{
		Ctx: ctx,
		DB:  db,
	}
}

func (c *SystemMessage) Create(data *model.SystemMessage) error {
	return c.DB.WithContext(c.Ctx).Create(data).Error
}

func (c *SystemMessage) Update(data *model.SystemMessage) error {
	return c.DB.WithContext(c.Ctx).Where("id = ?", data.ID).Updates(data).Error
}
