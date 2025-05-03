package repository

import (
	"context"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"

	"gorm.io/gorm"
)

type Message struct {
	Base[model.Message]
}

func NewMessageRepo(ctx context.Context, db *gorm.DB) *Message {
	return &Message{
		Base[model.Message]{
			Ctx: ctx,
			DB:  db,
		}}
}

func (m *Message) FindByContactID(req dto.ChatHistoryRequest, pager appx.Pager) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64

	query := m.DB.Model(&model.Message{})
	query = query.Where("from_wxid = ?", req.ContactID).Where("to_wxid = ?", req.Owner)
	if req.Keyword != "" {
		query = query.Where("content LIKE ?", "%"+req.Keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("id DESC").
		Offset(pager.OffSet).
		Limit(pager.PageSize).
		Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}
