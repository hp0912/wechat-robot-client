package repository

import (
	"context"
	"strings"
	"time"
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

func (m *Message) GetByID(id int64, preloads ...string) *model.Message {
	return m.takeOne(preloads, func(g *gorm.DB) *gorm.DB {
		return g.Where("id = ?", id)
	})
}

func (m *Message) GetByMsgID(msgId int64, preloads ...string) *model.Message {
	return m.takeOne(preloads, func(g *gorm.DB) *gorm.DB {
		return g.Where("msg_id = ?", msgId)
	})
}

func (m *Message) GetByContactID(req dto.ChatHistoryRequest, pager appx.Pager) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64
	query := m.DB.Model(&model.Message{})
	// 判断是群聊还是单聊，决定关联哪张表
	if strings.HasSuffix(req.ContactID, "@chatroom") {
		// 群聊，需要关联 chat_room_members 以获取发送者昵称和头像
		query = query.
			Joins("LEFT JOIN chat_room_members ON chat_room_members.wechat_id = messages.sender_wxid AND chat_room_members.chat_room_id = messages.from_wxid").
			Select("messages.*, chat_room_members.nickname AS sender_nickname, chat_room_members.avatar AS sender_avatar")
	} else {
		// 好友，需要关联 contacts 表
		query = query.
			Joins("LEFT JOIN contacts ON contacts.wechat_id = messages.sender_wxid").
			Select("messages.*, contacts.nickname AS sender_nickname, contacts.avatar AS sender_avatar")
	}
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

func (m *Message) GetYesterdayChatInfo(owner, chatRoomID string) ([]dto.ChatRoomSummary, error) {
	chatRoomSummary := []dto.ChatRoomSummary{}
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()
	query := m.DB.Model(&model.Message{})
	query = query.Select("count( 1 ) AS `member_chat_count`").
		Where("from_wxid = ?", chatRoomID).
		Where("to_wxid = ?", owner).
		Where("type < 10000").
		Where("created_at >= ?", yesterdayStartTimestamp).
		Where("created_at < ?", todayStartTimestamp).
		Group("sender_wxid")
	if err := query.Find(&chatRoomSummary).Error; err != nil {
		return nil, err
	}
	return chatRoomSummary, nil
}

func (m *Message) GetYesterdayChatRommRank(owner, chatRoomID string) ([]*dto.ChatRoomRank, error) {
	var chatRoomRank []*dto.ChatRoomRank
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()
	query := m.DB.Model(&model.Message{})
	query = query.Select("messages.sender_wxid", "chat_room_members.nickname", "count( 1 ) AS `count`").
		Joins("LEFT JOIN chat_room_members ON chat_room_members.wechat_id = messages.sender_wxid AND chat_room_members.chat_room_id = messages.from_wxid").
		Where("messages.from_wxid = ?", chatRoomID).
		Where("messages.to_wxid = ?", owner).
		Where("messages.type < 10000").
		Where("messages.created_at >= ?", yesterdayStartTimestamp).
		Where("messages.created_at < ?", todayStartTimestamp).
		Group("messages.sender_wxid").
		Order("`count` DESC")
	if err := query.Find(&chatRoomRank).Error; err != nil {
		return chatRoomRank, err
	}
	return chatRoomRank, nil
}

func (m *Message) GetLastWeekChatRommRank(owner, chatRoomID string) ([]*dto.ChatRoomRank, error) {
	var chatRoomRank []*dto.ChatRoomRank
	// 获取当前时间（周一）
	now := time.Now()
	// 计算上周一的零点时间
	// 当前是周一(weekday=1)，需要回退7天到上周一
	lastMondayStart := time.Date(now.Year(), now.Month(), now.Day()-7, 0, 0, 0, 0, now.Location())
	// 计算本周一的零点时间（上周日23:59:59的下一秒）
	thisWeekStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 转换为时间戳（秒）
	lastWeekStartTimestamp := lastMondayStart.Unix()
	thisWeekStartTimestamp := thisWeekStart.Unix()
	query := m.DB.Model(&model.Message{})
	query = query.Select("messages.sender_wxid", "chat_room_members.nickname", "count( 1 ) AS `count`").
		Joins("LEFT JOIN chat_room_members ON chat_room_members.wechat_id = messages.sender_wxid AND chat_room_members.chat_room_id = messages.from_wxid").
		Where("messages.from_wxid = ?", chatRoomID).
		Where("messages.to_wxid = ?", owner).
		Where("messages.type < 10000").
		Where("messages.created_at >= ?", lastWeekStartTimestamp).
		Where("messages.created_at < ?", thisWeekStartTimestamp).
		Group("messages.sender_wxid").
		Order("`count` DESC")
	if err := query.Find(&chatRoomRank).Error; err != nil {
		return chatRoomRank, err
	}
	return chatRoomRank, nil
}

func (m *Message) GetLastMonthChatRommRank(owner, chatRoomID string) ([]*dto.ChatRoomRank, error) {
	var chatRoomRank []*dto.ChatRoomRank
	// 获取当前时间（每月一号执行）
	now := time.Now()
	// 获取本月1号零点
	thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	// 获取上月1号零点
	lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
	// 转换为时间戳（秒）
	lastMonthStartTimestamp := lastMonthStart.Unix()
	thisMonthStartTimestamp := thisMonthStart.Unix()
	query := m.DB.Model(&model.Message{})
	query = query.Select("messages.sender_wxid", "chat_room_members.nickname", "count( 1 ) AS `count`").
		Joins("LEFT JOIN chat_room_members ON chat_room_members.wechat_id = messages.sender_wxid AND chat_room_members.chat_room_id = messages.from_wxid").
		Where("messages.from_wxid = ?", chatRoomID).
		Where("messages.to_wxid = ?", owner).
		Where("messages.type < 10000").
		Where("messages.created_at >= ?", lastMonthStartTimestamp).
		Where("messages.created_at < ?", thisMonthStartTimestamp).
		Group("messages.sender_wxid").
		Order("`count` DESC")
	if err := query.Find(&chatRoomRank).Error; err != nil {
		return chatRoomRank, err
	}
	return chatRoomRank, nil
}
