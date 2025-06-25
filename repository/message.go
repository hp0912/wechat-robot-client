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
	Ctx context.Context
	DB  *gorm.DB
}

func NewMessageRepo(ctx context.Context, db *gorm.DB) *Message {
	return &Message{
		Ctx: ctx,
		DB:  db,
	}
}

func (m *Message) GetByID(id int64) (*model.Message, error) {
	var message model.Message
	err := m.DB.WithContext(m.Ctx).Where("id = ?", id).First(&message).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (m *Message) GetByMsgID(msgId int64) (*model.Message, error) {
	var message model.Message
	err := m.DB.WithContext(m.Ctx).Where("msg_id = ?", msgId).First(&message).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (m *Message) GetByContactID(req dto.ChatHistoryRequest, pager appx.Pager) ([]*model.Message, int64, error) {
	var messages []*model.Message
	var total int64
	query := m.DB.WithContext(m.Ctx).Model(&model.Message{})
	// 判断是群聊还是单聊，决定关联哪张表
	if strings.HasSuffix(req.ContactID, "@chatroom") {
		// 群聊，需要关联 chat_room_members 以获取发送者昵称和头像
		query = query.
			Joins("LEFT JOIN chat_room_members ON chat_room_members.wechat_id = messages.sender_wxid AND chat_room_members.chat_room_id = messages.from_wxid").
			Select("messages.*, IF(chat_room_members.remark != '' AND chat_room_members.remark IS NOT NULL, chat_room_members.remark, chat_room_members.nickname) AS sender_nickname, chat_room_members.avatar AS sender_avatar")
	} else {
		// 好友，需要关联 contacts 表
		query = query.
			Joins("LEFT JOIN contacts ON contacts.wechat_id = messages.sender_wxid").
			Select("messages.*, IF(contacts.remark != '' AND contacts.remark IS NOT NULL, contacts.remark, contacts.nickname) AS sender_nickname, contacts.avatar AS sender_avatar")
	}
	query = query.Where("from_wxid = ?", req.ContactID)
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

func (m *Message) SetMessageIsInContext(message *model.Message) error {
	return m.DB.WithContext(m.Ctx).Where("id = ?", message.ID).Updates(&model.Message{IsAIContext: true}).Error
}

func (m *Message) GetFriendAIMessageContext(message *model.Message) ([]*model.Message, error) {
	var messages []*model.Message
	tenMinutesAgo := time.Now().Add(-10 * time.Minute).Unix()
	err := m.DB.WithContext(m.Ctx).Where("id < ?", message.ID).
		Where("from_wxid = ?", message.FromWxID).
		Where("created_at >= ?", tenMinutesAgo).
		Where("`type` in (1, 3) OR (`type` = 49 AND `app_msg_type` = 57)").
		Find(&messages).
		Order("id ASC").Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (m *Message) ResetFriendAIMessageContext(message *model.Message) error {
	tenMinutesAgo := time.Now().Add(-10 * time.Minute).Unix()
	return m.DB.WithContext(m.Ctx).Model(&model.Message{}).
		Where("id < ?", message.ID).
		Where("from_wxid = ?", message.FromWxID).
		Where("created_at >= ?", tenMinutesAgo).
		Updates(map[string]int{"is_ai_context": 0}).Error
}

func (m *Message) GetChatRoomAIMessageContext(message *model.Message) ([]*model.Message, error) {
	var messages []*model.Message
	tenMinutesAgo := time.Now().Add(-10 * time.Minute).Unix()
	err := m.DB.WithContext(m.Ctx).Where("id < ?", message.ID).
		Where("from_wxid = ?", message.FromWxID).
		Where("(sender_wxid = ? AND is_ai_context = 1) OR reply_wxid = ?", message.SenderWxID, message.SenderWxID).
		Where("created_at >= ?", tenMinutesAgo).
		Where("`type` in (1, 3) OR (`type` = 49 AND `app_msg_type` = 57)").
		Find(&messages).
		Order("id ASC").Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (m *Message) ResetChatRoomAIMessageContext(message *model.Message) error {
	tenMinutesAgo := time.Now().Add(-10 * time.Minute).Unix()
	return m.DB.WithContext(m.Ctx).Where("id < ?", message.ID).
		Where("from_wxid = ?", message.FromWxID).
		Where("(sender_wxid = ? AND is_ai_context = 1) OR reply_wxid = ?", message.SenderWxID, message.SenderWxID).
		Where("created_at >= ?", tenMinutesAgo).
		Updates(map[string]int{"is_ai_context": 0}).Error
}

func (m *Message) GetMessagesByTimeRange(self, chatRoomID string, startTime, endTime int64) ([]*dto.TextMessageItem, error) {
	var messages []*dto.TextMessageItem
	// APP消息类型
	appMsgList := []string{"57", "4", "5", "6"}
	// 这个查询子句抽出来写，方便后续扩展
	selectStr := `CASE
		WHEN messages.type = 49 THEN
	CASE
			WHEN EXTRACTVALUE ( messages.content, "/msg/appmsg/type" ) = '57' THEN
			EXTRACTVALUE ( messages.content, "/msg/appmsg/title" )
			WHEN EXTRACTVALUE ( messages.content, "/msg/appmsg/type" ) = '5' THEN
			CONCAT("网页分享消息，标题: ", EXTRACTVALUE (messages.content, "/msg/appmsg/title"), "，描述：", EXTRACTVALUE (messages.content, "/msg/appmsg/des"))
			WHEN EXTRACTVALUE ( messages.content, "/msg/appmsg/type" ) = '4' THEN
			CONCAT("网页分享消息，标题: ", EXTRACTVALUE (messages.content, "/msg/appmsg/title"), "，描述：", EXTRACTVALUE (messages.content, "/msg/appmsg/des"))
			WHEN EXTRACTVALUE ( messages.content, "/msg/appmsg/type" ) = '6' THEN
			CONCAT("文件消息，文件名: ", EXTRACTVALUE (messages.content, "/msg/appmsg/title"))

			ELSE EXTRACTVALUE ( messages.content, "/msg/appmsg/des" )
		END ELSE messages.content
	END`
	query := m.DB.WithContext(m.Ctx).Model(&model.Message{})
	query = query.Select("IF(chat_room_members.remark != '' AND chat_room_members.remark IS NOT NULL, chat_room_members.remark, chat_room_members.nickname) AS nickname", selectStr+" AS message", "messages.created_at").
		Joins("LEFT JOIN chat_room_members ON chat_room_members.wechat_id = messages.sender_wxid AND chat_room_members.chat_room_id = messages.from_wxid").
		Where("messages.from_wxid = ?", chatRoomID).
		Where(`(messages.type = 1 OR ( messages.type = 49 AND EXTRACTVALUE ( messages.content, "/msg/appmsg/type" ) IN (?) ))`, appMsgList).
		Where("messages.sender_wxid != ?", self).
		Where("messages.created_at >= ?", startTime).
		Where("messages.created_at < ?", endTime).
		Order("messages.created_at ASC")
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (m *Message) GetYesterdayChatInfo(chatRoomID string) ([]*dto.ChatRoomSummary, error) {
	var chatRoomSummary []*dto.ChatRoomSummary
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()
	query := m.DB.WithContext(m.Ctx).Model(&model.Message{})
	query = query.Select("count( 1 ) AS `message_count`").
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

func (m *Message) GetYesterdayChatRommRank(self, chatRoomID string) ([]*dto.ChatRoomRank, error) {
	var chatRoomRank []*dto.ChatRoomRank
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	// 获取昨天凌晨零点
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	// 转换为时间戳（秒）
	yesterdayStartTimestamp := yesterdayStart.Unix()
	todayStartTimestamp := todayStart.Unix()
	query := m.DB.WithContext(m.Ctx).Model(&model.Message{})
	query = query.Select("messages.sender_wxid", "IF(chat_room_members.remark != '' AND chat_room_members.remark IS NOT NULL, chat_room_members.remark, chat_room_members.nickname) AS chat_room_member_nickname", "count( 1 ) AS `count`").
		Joins("LEFT JOIN chat_room_members ON chat_room_members.wechat_id = messages.sender_wxid AND chat_room_members.chat_room_id = messages.from_wxid").
		Where("messages.from_wxid = ?", chatRoomID).
		Where("messages.type < 10000").
		Where("messages.sender_wxid != ?", self).
		Where("messages.created_at >= ?", yesterdayStartTimestamp).
		Where("messages.created_at < ?", todayStartTimestamp).
		Group("messages.sender_wxid, chat_room_member_nickname").
		Order("`count` DESC")
	if err := query.Find(&chatRoomRank).Error; err != nil {
		return chatRoomRank, err
	}
	return chatRoomRank, nil
}

func (m *Message) GetLastWeekChatRommRank(self, chatRoomID string) ([]*dto.ChatRoomRank, error) {
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
	query := m.DB.WithContext(m.Ctx).Model(&model.Message{})
	query = query.Select("messages.sender_wxid", "IF(chat_room_members.remark != '' AND chat_room_members.remark IS NOT NULL, chat_room_members.remark, chat_room_members.nickname) AS chat_room_member_nickname", "count( 1 ) AS `count`").
		Joins("LEFT JOIN chat_room_members ON chat_room_members.wechat_id = messages.sender_wxid AND chat_room_members.chat_room_id = messages.from_wxid").
		Where("messages.from_wxid = ?", chatRoomID).
		Where("messages.type < 10000").
		Where("messages.sender_wxid != ?", self).
		Where("messages.created_at >= ?", lastWeekStartTimestamp).
		Where("messages.created_at < ?", thisWeekStartTimestamp).
		Group("messages.sender_wxid, chat_room_member_nickname").
		Order("`count` DESC")
	if err := query.Find(&chatRoomRank).Error; err != nil {
		return chatRoomRank, err
	}
	return chatRoomRank, nil
}

func (m *Message) GetLastMonthChatRommRank(self, chatRoomID string) ([]*dto.ChatRoomRank, error) {
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
	query := m.DB.WithContext(m.Ctx).Model(&model.Message{})
	query = query.Select("messages.sender_wxid", "IF(chat_room_members.remark != '' AND chat_room_members.remark IS NOT NULL, chat_room_members.remark, chat_room_members.nickname) AS chat_room_member_nickname", "count( 1 ) AS `count`").
		Joins("LEFT JOIN chat_room_members ON chat_room_members.wechat_id = messages.sender_wxid AND chat_room_members.chat_room_id = messages.from_wxid").
		Where("messages.from_wxid = ?", chatRoomID).
		Where("messages.type < 10000").
		Where("messages.sender_wxid != ?", self).
		Where("messages.created_at >= ?", lastMonthStartTimestamp).
		Where("messages.created_at < ?", thisMonthStartTimestamp).
		Group("messages.sender_wxid, chat_room_member_nickname").
		Order("`count` DESC")
	if err := query.Find(&chatRoomRank).Error; err != nil {
		return chatRoomRank, err
	}
	return chatRoomRank, nil
}

func (m *Message) Create(data *model.Message) error {
	return m.DB.WithContext(m.Ctx).Create(data).Error
}

func (m *Message) Update(data *model.Message) error {
	return m.DB.WithContext(m.Ctx).Where("id = ?", data.ID).Updates(data).Error
}
