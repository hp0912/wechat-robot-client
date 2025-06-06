package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type GlobalSettings struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewGlobalSettingsRepo(ctx context.Context, db *gorm.DB) *GlobalSettings {
	return &GlobalSettings{
		Ctx: ctx,
		DB:  db,
	}
}

func (respo *GlobalSettings) GetGlobalSettings() (*model.GlobalSettings, error) {
	var globalSettings model.GlobalSettings
	err := respo.DB.WithContext(respo.Ctx).First(&globalSettings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &globalSettings, nil
}

func (respo *GlobalSettings) GetRandomOne() (*model.GlobalSettings, error) {
	var globalSettings model.GlobalSettings
	err := respo.DB.WithContext(respo.Ctx).First(&globalSettings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &globalSettings, nil
}

func (respo *GlobalSettings) Create(data *model.GlobalSettings) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *GlobalSettings) Update(data *model.GlobalSettings) error {
	return respo.DB.WithContext(respo.Ctx).Where("id = ?", data.ID).Updates(data).Error
}
