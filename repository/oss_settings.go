package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type OSSSettings struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewOSSSettingsRepo(ctx context.Context, db *gorm.DB) *OSSSettings {
	return &OSSSettings{
		Ctx: ctx,
		DB:  db,
	}
}

func (respo *OSSSettings) GetOSSSettings() (*model.OSSSettings, error) {
	var ossSettings model.OSSSettings
	err := respo.DB.WithContext(respo.Ctx).First(&ossSettings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ossSettings, nil
}

func (respo *OSSSettings) Create(data *model.OSSSettings) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *OSSSettings) Update(data *model.OSSSettings) error {
	return respo.DB.WithContext(respo.Ctx).Updates(data).Error
}
