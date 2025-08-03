package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type Moment struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewMomentRepo(ctx context.Context, db *gorm.DB) *Moment {
	return &Moment{
		Ctx: ctx,
		DB:  db,
	}
}

func (respo *Moment) GetMomentByID(momentID int64) (*model.Moment, error) {
	var moment model.Moment
	err := respo.DB.WithContext(respo.Ctx).Where("moment_id = ?", momentID).First(&moment).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &moment, nil
}

func (respo *Moment) Create(data *model.Moment) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *Moment) Update(data *model.Moment) error {
	return respo.DB.WithContext(respo.Ctx).Updates(data).Error
}
