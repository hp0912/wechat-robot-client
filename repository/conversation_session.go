package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type ConversationSession struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewConversationSessionRepo(ctx context.Context, db *gorm.DB) *ConversationSession {
	return &ConversationSession{Ctx: ctx, DB: db}
}

func (r *ConversationSession) Create(session *model.ConversationSession) error {
	now := time.Now().Unix()
	session.CreatedAt = now
	session.UpdatedAt = now
	return r.DB.WithContext(r.Ctx).Create(session).Error
}

func (r *ConversationSession) Update(session *model.ConversationSession) error {
	session.UpdatedAt = time.Now().Unix()
	return r.DB.WithContext(r.Ctx).Save(session).Error
}

// GetActiveSession 获取某个联系人当前活跃的会话
func (r *ConversationSession) GetActiveSession(contactWxID, chatRoomID string) (*model.ConversationSession, error) {
	var session model.ConversationSession
	query := r.DB.WithContext(r.Ctx).Where("is_active = ?", true)
	if chatRoomID != "" {
		query = query.Where("chat_room_id = ? AND contact_wxid = ?", chatRoomID, contactWxID)
	} else {
		query = query.Where("contact_wxid = ?", contactWxID)
	}
	err := query.Order("id DESC").First(&session).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &session, err
}

// GetLatestSummary 获取某个联系人上一次的对话摘要
func (r *ConversationSession) GetLatestSummary(contactWxID, chatRoomID string) (string, error) {
	var session model.ConversationSession
	query := r.DB.WithContext(r.Ctx).
		Where("is_active = ? AND summary != ''", false)
	if chatRoomID != "" {
		query = query.Where("chat_room_id = ? AND contact_wxid = ?", chatRoomID, contactWxID)
	} else {
		query = query.Where("contact_wxid = ?", contactWxID)
	}
	err := query.Order("id DESC").First(&session).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	return session.Summary, err
}

// CloseExpiredSessions 关闭超过指定分钟数不活跃的会话
func (r *ConversationSession) CloseExpiredSessions(inactiveMinutes int) ([]*model.ConversationSession, error) {
	threshold := time.Now().Add(-time.Duration(inactiveMinutes) * time.Minute).Unix()
	var sessions []*model.ConversationSession
	err := r.DB.WithContext(r.Ctx).
		Where("is_active = ? AND last_active_at < ?", true, threshold).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	if len(sessions) > 0 {
		r.DB.WithContext(r.Ctx).
			Model(&model.ConversationSession{}).
			Where("is_active = ? AND last_active_at < ?", true, threshold).
			Update("is_active", false)
	}
	return sessions, nil
}
