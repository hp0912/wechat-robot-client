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
		query = query.Where("contact_wxid = ? AND chat_room_id = ''", contactWxID)
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
		query = query.Where("contact_wxid = ? AND chat_room_id = ''", contactWxID)
	}
	err := query.Order("id DESC").First(&session).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	return session.Summary, err
}

// CloseExpiredSessions 关闭超过指定分钟数不活跃的会话
// 先原子性地标记关闭，再查出需要摘要的会话，避免竞态条件
func (r *ConversationSession) CloseExpiredSessions(inactiveMinutes int) ([]*model.ConversationSession, error) {
	threshold := time.Now().Add(-time.Duration(inactiveMinutes) * time.Minute).Unix()

	// 先原子更新，返回受影响的 ID 列表
	var expiredIDs []int64
	err := r.DB.WithContext(r.Ctx).
		Model(&model.ConversationSession{}).
		Where("is_active = ? AND last_active_at < ?", true, threshold).
		Pluck("id", &expiredIDs).Error
	if err != nil || len(expiredIDs) == 0 {
		return nil, err
	}

	r.DB.WithContext(r.Ctx).
		Model(&model.ConversationSession{}).
		Where("id IN ? AND is_active = ?", expiredIDs, true).
		Update("is_active", false)

	var sessions []*model.ConversationSession
	err = r.DB.WithContext(r.Ctx).
		Where("id IN ?", expiredIDs).
		Find(&sessions).Error
	return sessions, err
}

// CloseExpiredPrivateSessions 关闭私聊中超过指定分钟数不活跃的会话
func (r *ConversationSession) CloseExpiredPrivateSessions(inactiveMinutes int) ([]*model.ConversationSession, error) {
	return r.closeExpiredByScope(inactiveMinutes, "chat_room_id = ''")
}

// CloseExpiredGroupSessions 关闭群聊中超过指定分钟数不活跃的会话
func (r *ConversationSession) CloseExpiredGroupSessions(inactiveMinutes int) ([]*model.ConversationSession, error) {
	return r.closeExpiredByScope(inactiveMinutes, "chat_room_id != ''")
}

func (r *ConversationSession) closeExpiredByScope(inactiveMinutes int, scopeCondition string) ([]*model.ConversationSession, error) {
	threshold := time.Now().Add(-time.Duration(inactiveMinutes) * time.Minute).Unix()

	var expiredIDs []int64
	err := r.DB.WithContext(r.Ctx).
		Model(&model.ConversationSession{}).
		Where("is_active = ? AND last_active_at < ?", true, threshold).
		Where(scopeCondition).
		Pluck("id", &expiredIDs).Error
	if err != nil || len(expiredIDs) == 0 {
		return nil, err
	}

	r.DB.WithContext(r.Ctx).
		Model(&model.ConversationSession{}).
		Where("id IN ? AND is_active = ?", expiredIDs, true).
		Update("is_active", false)

	var sessions []*model.ConversationSession
	err = r.DB.WithContext(r.Ctx).
		Where("id IN ?", expiredIDs).
		Find(&sessions).Error
	return sessions, err
}
