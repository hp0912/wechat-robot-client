package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type GlobalSettings struct {
	Base[model.GlobalSettings]
}

func NewGlobalSettingsRepo(ctx context.Context, db *gorm.DB) *GlobalSettings {
	return &GlobalSettings{
		Base[model.GlobalSettings]{
			Ctx: ctx,
			DB:  db,
		}}
}

func (respo *GlobalSettings) GetByOwner(owner string, preloads ...string) *model.GlobalSettings {
	return respo.takeOne(preloads, func(g *gorm.DB) *gorm.DB {
		return g.Where("owner = ?", owner)
	})
}

func (respo *GlobalSettings) GetRandomOne(preloads ...string) *model.GlobalSettings {
	return respo.takeOne(preloads)
}
