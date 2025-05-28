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

func (c *ChatRoomMember) GetByChatRoomID(req dto.ChatRoomMemberRequest, pager appx.Pager, preloads ...string) ([]*model.ChatRoomMember, int64, error) {
	var chatRoomMembers []*model.ChatRoomMember
	var total int64

	query := c.DB.Model(&model.ChatRoomMember{})
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	query = query.Where("chat_room_id = ?", req.ChatRoomID)
	if req.Keyword != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Keyword+"%").
			Or("alias LIKE ?", "%"+req.Keyword+"%").
			Or("wechat_id LIKE ?", "%"+req.Keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("last_active_at DESC").Order("id DESC")
	if err := query.Offset(pager.OffSet).Limit(pager.PageSize).Find(&chatRoomMembers).Error; err != nil {
		return nil, 0, err
	}

	return chatRoomMembers, total, nil
}

// 当前群总人数
func (c *ChatRoomMember) GetChatRoomMemberCount(chatRoomID string) (int64, error) {
	var total int64
	query := c.DB.Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ?", chatRoomID).Where("leaved_at IS NULL")
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// 昨天入群人数
func (c *ChatRoomMember) GetYesterdayJoinCount(chatRoomID string) (int64, error) {
	var total int64
	query := c.DB.Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ?", chatRoomID).
		Where("joined_at >= ?", 0).
		Where("joined_at < ?", 0)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
