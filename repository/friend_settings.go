package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type FriendSettings struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewFriendSettingsRepo(ctx context.Context, db *gorm.DB) *FriendSettings {
	return &FriendSettings{
		Ctx: ctx,
		DB:  db,
	}
}

func (respo *FriendSettings) GetByOwner(owner, contactID string) (*model.FriendSettings, error) {
	var friendSettings model.FriendSettings
	err := respo.DB.WithContext(respo.Ctx).Where("owner = ? AND wechat_id = ?", owner, contactID).First(&friendSettings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &friendSettings, nil
}

func (respo *FriendSettings) Create(data *model.FriendSettings) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *FriendSettings) Update(data *model.FriendSettings) error {
	return respo.DB.WithContext(respo.Ctx).Updates(data).Error
}
