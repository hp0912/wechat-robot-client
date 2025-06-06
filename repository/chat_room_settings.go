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

func (respo *ChatRoomSettings) GetChatRoomSettings(chatRoomID string) (*model.ChatRoomSettings, error) {
	var chatRoomSettings model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("chat_room_id = ?", chatRoomID).First(&chatRoomSettings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &chatRoomSettings, nil
}

func (respo *ChatRoomSettings) GetAllEnableGoodMorning() ([]*model.ChatRoomSettings, error) {
	var chatRoomSettings []*model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("morning_enabled = ?", 1).Find(&chatRoomSettings).Error
	if err != nil {
		return nil, err
	}
	return chatRoomSettings, nil
}

func (respo *ChatRoomSettings) GetAllEnableNews() ([]*model.ChatRoomSettings, error) {
	var chatRoomSettings []*model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("news_enabled = ?", 1).Find(&chatRoomSettings).Error
	if err != nil {
		return nil, err
	}
	return chatRoomSettings, nil
}

func (respo *ChatRoomSettings) GetAllEnableChatRank() ([]*model.ChatRoomSettings, error) {
	var chatRoomSettings []*model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("chat_room_ranking_enabled = ?", 1).Find(&chatRoomSettings).Error
	if err != nil {
		return nil, err
	}
	return chatRoomSettings, nil
}

func (respo *ChatRoomSettings) GetAllEnableAISummary() ([]*model.ChatRoomSettings, error) {
	var chatRoomSettings []*model.ChatRoomSettings
	err := respo.DB.WithContext(respo.Ctx).Where("chat_room_summary_enabled = ?", 1).Find(&chatRoomSettings).Error
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
