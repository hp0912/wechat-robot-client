package repository

import (
	"context"
	"time"
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
	query = query.Where("chat_room_id = ?", req.ChatRoomID).Where("owner = ?", req.Owner)
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
func (c *ChatRoomMember) GetChatRoomMemberCount(owner, chatRoomID string) (int64, error) {
	var total int64
	query := c.DB.Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ?", chatRoomID).Where("owner = ?", owner).Where("leaved_at IS NULL")
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// 昨天入群人数
func (c *ChatRoomMember) GetYesterdayJoinCount(owner, chatRoomID string) (int64, error) {
	var total int64
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()
	query := c.DB.Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ?", chatRoomID).
		Where("owner = ?", owner).
		Where("joined_at >= ?", yesterdayStartTimestamp).
		Where("joined_at < ?", todayStartTimestamp)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// 昨天离群人数
func (c *ChatRoomMember) GetYesterdayLeaveCount(owner, chatRoomID string) (int64, error) {
	var total int64
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()
	query := c.DB.Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ?", chatRoomID).
		Where("owner = ?", owner).
		Where("leaved_at >= ?", yesterdayStartTimestamp).
		Where("leaved_at < ?", todayStartTimestamp)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
