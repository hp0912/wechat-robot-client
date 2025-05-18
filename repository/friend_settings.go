package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type FriendSettings struct {
	Base[model.FriendSettings]
}

func NewFriendSettingsRepo(ctx context.Context, db *gorm.DB) *FriendSettings {
	return &FriendSettings{
		Base[model.FriendSettings]{
			Ctx: ctx,
			DB:  db,
		}}
}

func (respo *FriendSettings) GetByOwner(owner, contactID string, preloads ...string) *model.FriendSettings {
	return respo.takeOne(preloads, func(g *gorm.DB) *gorm.DB {
		return g.Where("owner = ?", owner).Where("wechat_id = ?", contactID)
	})
}
