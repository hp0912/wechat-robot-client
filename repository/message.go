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

func (m *Message) GetYesterdayChatInfo(chatRoomID string) ([]dto.ChatRoomSummary, error) {
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
		Where("type < 10000").
		Where("created_at >= ?", yesterdayStartTimestamp).
		Where("created_at < ?", todayStartTimestamp).
		Group("sender_wxid")

	if err := query.Find(&chatRoomSummary).Error; err != nil {
		return nil, err
	}
	return chatRoomSummary, nil
}
