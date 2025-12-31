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
	Ctx context.Context
	DB  *gorm.DB
}

func NewChatRoomMemberRepo(ctx context.Context, db *gorm.DB) *ChatRoomMember {
	return &ChatRoomMember{
		Ctx: ctx,
		DB:  db,
	}
}

func (c *ChatRoomMember) GetByID(id int64) (*model.ChatRoomMember, error) {
	var chatRoomMember model.ChatRoomMember
	err := c.DB.WithContext(c.Ctx).Where("id = ?", id).First(&chatRoomMember).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &chatRoomMember, nil
}

func (c *ChatRoomMember) GetByChatRoomID(req dto.ChatRoomMemberListRequest, pager appx.Pager) ([]*model.ChatRoomMember, int64, error) {
	var chatRoomMembers []*model.ChatRoomMember
	var total int64
	query := c.DB.WithContext(c.Ctx).Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ?", req.ChatRoomID)
	if req.Keyword != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Keyword+"%").Or("remark LIKE ?", "%"+req.Keyword+"%")
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

func (c *ChatRoomMember) GetNotLeftMemberByChatRoomID(req dto.ChatRoomMemberListRequest) ([]*model.ChatRoomMember, error) {
	var chatRoomMembers []*model.ChatRoomMember
	query := c.DB.WithContext(c.Ctx).Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ? AND is_leaved = 0", req.ChatRoomID)
	if req.Keyword != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Keyword+"%").Or("remark LIKE ?", "%"+req.Keyword+"%")
	}
	if err := query.Find(&chatRoomMembers).Error; err != nil {
		return nil, err
	}
	return chatRoomMembers, nil
}

func (c *ChatRoomMember) GetChatRoomMember(chatRoomID, wechatID string) (*model.ChatRoomMember, error) {
	var chatRoomMember model.ChatRoomMember
	err := c.DB.WithContext(c.Ctx).Where("chat_room_id = ? AND wechat_id = ?", chatRoomID, wechatID).First(&chatRoomMember).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &chatRoomMember, nil
}

func (c *ChatRoomMember) GetChatRoomMemberByWeChatIDs(chatRoomID string, wechatIDs []string) ([]*model.ChatRoomMember, error) {
	var chatRoomMembers []*model.ChatRoomMember
	err := c.DB.WithContext(c.Ctx).Where("chat_room_id = ? AND wechat_id IN ?", chatRoomID, wechatIDs).Find(&chatRoomMembers).Error
	if err != nil {
		return nil, err
	}
	return chatRoomMembers, nil
}

// GetChatRoomMembers 获取未退出群聊的成员
func (c *ChatRoomMember) GetChatRoomMembers(chatRoomID string) ([]*model.ChatRoomMember, error) {
	var chatRoomMembers []*model.ChatRoomMember
	err := c.DB.WithContext(c.Ctx).Where("chat_room_id = ? AND leaved_at IS NULL", chatRoomID).Find(&chatRoomMembers).Error
	if err != nil {
		return nil, err
	}
	return chatRoomMembers, nil
}

// 当前群总人数
func (c *ChatRoomMember) GetChatRoomMemberCount(chatRoomID string) (int64, error) {
	var total int64
	query := c.DB.WithContext(c.Ctx).Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ?", chatRoomID).Where("leaved_at IS NULL")
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// 昨天入群人数
func (c *ChatRoomMember) GetYesterdayJoinCount(chatRoomID string) (int64, error) {
	var total int64
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()
	query := c.DB.WithContext(c.Ctx).Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ?", chatRoomID).
		Where("joined_at >= ?", yesterdayStartTimestamp).
		Where("joined_at < ?", todayStartTimestamp)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// 昨天离群人数
func (c *ChatRoomMember) GetYesterdayLeaveCount(chatRoomID string) (int64, error) {
	var total int64
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()
	query := c.DB.WithContext(c.Ctx).Model(&model.ChatRoomMember{})
	query = query.Where("chat_room_id = ?", chatRoomID).
		Where("leaved_at >= ?", yesterdayStartTimestamp).
		Where("leaved_at < ?", todayStartTimestamp)
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (c *ChatRoomMember) Create(data *model.ChatRoomMember) error {
	return c.DB.WithContext(c.Ctx).Create(data).Error
}

func (c *ChatRoomMember) Update(data *model.ChatRoomMember) error {
	return c.DB.WithContext(c.Ctx).Where("id = ?", data.ID).Updates(data).Error
}

func (c *ChatRoomMember) UpdateByID(id int64, data map[string]any) error {
	return c.DB.WithContext(c.Ctx).Model(&model.ChatRoomMember{}).Where("id = ?", id).Updates(data).Error
}

func (c *ChatRoomMember) DeleteChatRoomMembers(memberIDs []string) error {
	return c.DB.WithContext(c.Ctx).Model(&model.ChatRoomMember{}).
		Where("wechat_id IN (?)", memberIDs).
		Update("is_leaved", 1).
		Update("leaved_at", time.Now().Unix()).
		Error
}
