package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type MomentSettings struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewMomentSettingsRepo(ctx context.Context, db *gorm.DB) *MomentSettings {
	return &MomentSettings{
		Ctx: ctx,
		DB:  db,
	}
}

func (respo *MomentSettings) GetMomentSettings() (*model.MomentSettings, error) {
	var momentSettings model.MomentSettings
	err := respo.DB.WithContext(respo.Ctx).First(&momentSettings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &momentSettings, nil
}

func (respo *MomentSettings) Create(data *model.MomentSettings) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *MomentSettings) Update(data *model.MomentSettings) error {
	return respo.DB.WithContext(respo.Ctx).Updates(data).Error
}
