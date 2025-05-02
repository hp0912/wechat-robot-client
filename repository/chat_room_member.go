package repository

import (
	"context"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"

	"gorm.io/gorm"
)

type ChatRoomMember struct {
	Base[model.ChatRoomMember]
}

func NewChatRoomMemberRepo(ctx context.Context, db *gorm.DB) *ChatRoomMember {
	return &ChatRoomMember{
		Base[model.ChatRoomMember]{
			Ctx: ctx,
			DB:  db,
		}}
}

func (c *ChatRoomMember) FindByChatRoomID(req dto.ChatRoomMemberRequest, pager appx.Pager, preloads ...string) ([]*model.ChatRoomMember, int64, error) {
	query := c.DB.Model(&model.ChatRoomMember{})
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	query = query.Where("chat_room_id = ?", req.ChatRoomID)
	if req.Keyword != "" {
		query = query.Where("nickname LIKE ?", req.Keyword+"%").
			Or("alias LIKE ?", req.Keyword+"%").
			Or("wechat_id LIKE ?", req.Keyword+"%")
	}
	query = query.Order("last_active_at DESC").Order("id DESC")
	var chatRoomMembers []*model.ChatRoomMember
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(pager.OffSet).Limit(pager.PageSize).Find(&chatRoomMembers).Error; err != nil {
		return nil, 0, err
	}
	return chatRoomMembers, total, nil
}
