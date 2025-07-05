package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type SystemMessage struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewSystemMessageRepo(ctx context.Context, db *gorm.DB) *SystemMessage {
	return &SystemMessage{
		Ctx: ctx,
		DB:  db,
	}
}

func (respo *SystemMessage) GetByID(id int64) (*model.SystemMessage, error) {
	var message model.SystemMessage
	err := respo.DB.WithContext(respo.Ctx).Where("id = ?", id).First(&message).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (respo *SystemMessage) GetRecentMessages(days int) ([]*model.SystemMessage, error) {
	var messages []*model.SystemMessage
	startTime := time.Now().AddDate(0, 0, -days).Unix()
	err := respo.DB.WithContext(respo.Ctx).Where("created_at >= ?", startTime).Order("created_at DESC").Find(&messages).Error
	return messages, err
}

// GetRecentMonthMessages 获取最近一个月的系统消息
func (respo *SystemMessage) GetRecentMonthMessages() ([]*model.SystemMessage, error) {
	return respo.GetRecentMessages(30)
}

// GetRecentMessagesByType 获取最近指定天数的特定类型系统消息
func (respo *SystemMessage) GetRecentMessagesByType(days int, msgType model.SystemMessageType) ([]*model.SystemMessage, error) {
	var messages []*model.SystemMessage
	startTime := time.Now().AddDate(0, 0, -days).Unix()
	err := respo.DB.WithContext(respo.Ctx).Where("created_at >= ? AND type = ?", startTime, msgType).Order("created_at DESC").Find(&messages).Error
	return messages, err
}

// GetUnreadMessages 获取未读的系统消息
func (respo *SystemMessage) GetUnreadMessages() ([]*model.SystemMessage, error) {
	var messages []*model.SystemMessage
	err := respo.DB.WithContext(respo.Ctx).Where("is_read = ?", false).Order("created_at DESC").Find(&messages).Error
	return messages, err
}

// MarkAsRead 标记消息为已读
func (respo *SystemMessage) MarkAsRead(id int64) error {
	return respo.DB.WithContext(respo.Ctx).Model(&model.SystemMessage{}).Where("id = ?", id).Update("is_read", true).Error
}

func (respo *SystemMessage) Create(data *model.SystemMessage) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *SystemMessage) Update(data *model.SystemMessage) error {
	return respo.DB.WithContext(respo.Ctx).Where("id = ?", data.ID).Updates(data).Error
}
