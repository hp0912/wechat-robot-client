package repository

import (
	"context"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type ChatRoomSettings struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewChatRoomSettingsRepo(ctx context.Context, db *gorm.DB) *ChatRoomSettings {
	return &ChatRoomSettings{
		Ctx: ctx,
		DB:  db,
	}
}

func (respo *ChatRoomSettings) GetByOwner(owner, chatRoomID string) (*model.ChatRoomSettings, error) {
	var chatRoomSettings model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("owner = ? AND chat_room_id = ?", owner, chatRoomID).First(&chatRoomSettings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &chatRoomSettings, nil
}

func (respo *ChatRoomSettings) GetAllEnableGoodMorning(owner string) ([]*model.ChatRoomSettings, error) {
	var chatRoomSettings []*model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("owner = ? AND morning_enabled = ?", owner, 1).Find(&chatRoomSettings).Error
	if err != nil {
		return nil, err
	}
	return chatRoomSettings, nil
}

func (respo *ChatRoomSettings) GetAllEnableNews(owner string) ([]*model.ChatRoomSettings, error) {
	var chatRoomSettings []*model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("owner = ? AND news_enabled = ?", owner, 1).Find(&chatRoomSettings).Error
	if err != nil {
		return nil, err
	}
	return chatRoomSettings, nil
}

func (respo *ChatRoomSettings) GetAllEnableChatRank(owner string) ([]*model.ChatRoomSettings, error) {
	var chatRoomSettings []*model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("owner = ? AND chat_room_ranking_enabled = ?", owner, 1).Find(&chatRoomSettings).Error
	if err != nil {
		return nil, err
	}
	return chatRoomSettings, nil
}

func (respo *ChatRoomSettings) GetAllEnableAISummary(owner string) ([]*model.ChatRoomSettings, error) {
	var chatRoomSettings []*model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("owner = ? AND chat_room_summary_enabled = ?", owner, 1).Find(&chatRoomSettings).Error
	if err != nil {
		return nil, err
	}
	return chatRoomSettings, nil
}

func (respo *ChatRoomSettings) Create(data *model.ChatRoomSettings) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *ChatRoomSettings) Update(data *model.ChatRoomSettings) error {
	return respo.DB.WithContext(respo.Ctx).Where("id = ?", data.ID).Updates(data).Error
}
