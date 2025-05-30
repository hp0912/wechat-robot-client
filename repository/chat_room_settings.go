package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type ChatRoomSettings struct {
	Base[model.ChatRoomSettings]
}

func NewChatRoomSettingsRepo(ctx context.Context, db *gorm.DB) *ChatRoomSettings {
	return &ChatRoomSettings{
		Base[model.ChatRoomSettings]{
			Ctx: ctx,
			DB:  db,
		}}
}

func (respo *ChatRoomSettings) GetByOwner(owner, chatRoomID string, preloads ...string) *model.ChatRoomSettings {
	return respo.takeOne(preloads, func(g *gorm.DB) *gorm.DB {
		return g.Where("owner = ?", owner).Where("chat_room_id = ?", chatRoomID)
	})
}

func (respo *ChatRoomSettings) GetAllEnableGoodMorning(owner string, preloads ...string) []*model.ChatRoomSettings {
	return respo.ListByWhere(preloads, where{
		"owner":           owner,
		"morning_enabled": 1,
	})
}

func (respo *ChatRoomSettings) GetAllEnableNews(owner string, preloads ...string) []*model.ChatRoomSettings {
	return respo.ListByWhere(preloads, where{
		"owner":        owner,
		"news_enabled": 1,
	})
}

func (respo *ChatRoomSettings) GetAllEnableChatRank(owner string, preloads ...string) []*model.ChatRoomSettings {
	return respo.ListByWhere(preloads, where{
		"owner":                     owner,
		"chat_room_ranking_enabled": 1,
	})
}

func (respo *ChatRoomSettings) GetAllEnableAISummary(owner string, preloads ...string) []*model.ChatRoomSettings {
	return respo.ListByWhere(preloads, where{
		"owner":                     owner,
		"chat_room_summary_enabled": 1,
	})
}
