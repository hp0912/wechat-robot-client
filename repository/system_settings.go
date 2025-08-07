package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type SystemSettings struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewSystemSettingsRepo(ctx context.Context, db *gorm.DB) *SystemSettings {
	return &SystemSettings{
		Ctx: ctx,
		DB:  db,
	}
}

func (respo *SystemSettings) GetMomentSettings() (*model.SystemSettings, error) {
	var systemSettings model.SystemSettings
	err := respo.DB.WithContext(respo.Ctx).First(&systemSettings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &systemSettings, nil
}

func (respo *SystemSettings) Create(data *model.SystemSettings) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *SystemSettings) Update(data *model.SystemSettings) error {
	return respo.DB.WithContext(respo.Ctx).Updates(data).Error
}
