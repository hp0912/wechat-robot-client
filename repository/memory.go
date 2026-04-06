package repository

import (
	"context"
	"strings"
	"time"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type Memory struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewMemoryRepo(ctx context.Context, db *gorm.DB) *Memory {
	return &Memory{Ctx: ctx, DB: db}
}

func (r *Memory) Create(memory *model.Memory) error {
	now := time.Now().Unix()
	memory.CreatedAt = now
	memory.UpdatedAt = now
	return r.DB.WithContext(r.Ctx).Create(memory).Error
}

func (r *Memory) Update(memory *model.Memory) error {
	memory.UpdatedAt = time.Now().Unix()
	return r.DB.WithContext(r.Ctx).Save(memory).Error
}

func (r *Memory) Delete(id int64) error {
	return r.DB.WithContext(r.Ctx).Delete(&model.Memory{}, id).Error
}

func (r *Memory) GetByID(id int64) (*model.Memory, error) {
	var memory model.Memory
	err := r.DB.WithContext(r.Ctx).Where("id = ?", id).First(&memory).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &memory, err
}

// GetByIDs 批量获取记忆
func (r *Memory) GetByIDs(ids []int64) ([]*model.Memory, error) {
	var memories []*model.Memory
	err := r.DB.WithContext(r.Ctx).
		Where("id IN ?", ids).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix()).
		Find(&memories).Error
	return memories, err
}

// GetByWxID 获取某人的所有记忆（全局），按重要性排序
func (r *Memory) GetByWxID(wxID string, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	err := r.DB.WithContext(r.Ctx).
		Where("wx_id = ? AND chat_room_id = ''", wxID).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix()).
		Order("importance DESC, updated_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// GetByWxIDAndChatRoom 获取某人在某群内的记忆
func (r *Memory) GetByWxIDAndChatRoom(wxID, chatRoomID string, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	err := r.DB.WithContext(r.Ctx).
		Where("wx_id = ? AND chat_room_id = ?", wxID, chatRoomID).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix()).
		Order("importance DESC, updated_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// GetByWxIDAllScopes 获取某人在所有作用域的记忆（全局 + 所有群）
func (r *Memory) GetByWxIDAllScopes(wxID string, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	err := r.DB.WithContext(r.Ctx).
		Where("wx_id = ?", wxID).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix()).
		Order("importance DESC, updated_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// GetByChatRoom 获取群级别记忆（wx_id 为空）
func (r *Memory) GetByChatRoom(chatRoomID string, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	err := r.DB.WithContext(r.Ctx).
		Where("wx_id = '' AND chat_room_id = ?", chatRoomID).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix()).
		Order("importance DESC, updated_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// ListByScope 获取某个会话视角下的候选记忆（全局个人 + 群内个人 + 群级别）。
func (r *Memory) ListByScope(wxID, chatRoomID string, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	query := r.scopeQuery(wxID, chatRoomID).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix())
	err := query.
		Order("importance DESC, updated_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// ListDueMemories 获取适合主动跟进的提醒类记忆。
func (r *Memory) ListDueMemories(wxID, chatRoomID string, until, cooldownBefore int64, limit int) ([]*model.Memory, error) {
	now := time.Now().Unix()
	var memories []*model.Memory
	query := r.scopeQuery(wxID, chatRoomID).
		Where("expire_at = 0 OR expire_at > ?", now).
		Where("last_access_at = 0 OR last_access_at < ?", cooldownBefore).
		Where("((reminder_at > 0 AND reminder_at <= ?) OR (category = ? AND happened_at > 0 AND happened_at >= ? AND happened_at <= ?))",
			until,
			model.MemoryCategoryEvent,
			now-12*60*60,
			until,
		)
	err := query.
		Order("importance DESC, reminder_at ASC, happened_at ASC, updated_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// GetAllByWxIDForProfile 获取某人的所有记忆，用于生成画像摘要
func (r *Memory) GetAllByWxIDForProfile(wxID, chatRoomID string) ([]*model.Memory, error) {
	var memories []*model.Memory
	query := r.DB.WithContext(r.Ctx).
		Where("wx_id = ?", wxID).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix())
	if chatRoomID != "" {
		query = query.Where("chat_room_id = '' OR chat_room_id = ?", chatRoomID)
	} else {
		query = query.Where("chat_room_id = ''")
	}
	err := query.Order("importance DESC, updated_at DESC").
		Limit(100).
		Find(&memories).Error
	return memories, err
}

// GetDistinctWxIDs 获取有记忆的所有用户 wxID 列表
func (r *Memory) GetDistinctWxIDs() ([]string, error) {
	var wxIDs []string
	err := r.DB.WithContext(r.Ctx).
		Model(&model.Memory{}).
		Where("wx_id != ''").
		Distinct("wx_id").
		Pluck("wx_id", &wxIDs).Error
	return wxIDs, err
}

// GetDistinctChatRoomsByWxID 获取某用户有群内记忆的所有群 ID
func (r *Memory) GetDistinctChatRoomsByWxID(wxID string) ([]string, error) {
	var chatRoomIDs []string
	err := r.DB.WithContext(r.Ctx).
		Model(&model.Memory{}).
		Where("wx_id = ? AND chat_room_id != ''", wxID).
		Distinct("chat_room_id").
		Pluck("chat_room_id", &chatRoomIDs).Error
	return chatRoomIDs, err
}

// IncrementAccessCount 增加访问计数
func (r *Memory) IncrementAccessCount(ids []int64) error {
	return r.DB.WithContext(r.Ctx).
		Model(&model.Memory{}).
		Where("id IN ?", ids).
		Updates(map[string]any{
			"access_count":   gorm.Expr("access_count + 1"),
			"last_access_at": time.Now().Unix(),
		}).Error
}

// DecayMemories 衰减长期未访问的记忆重要性
// 跳过带有重要关系标签或高权重关系类型的核心记忆，避免恋人/家人等信息被误衰减
func (r *Memory) DecayMemories(inactiveDays int) error {
	threshold := time.Now().AddDate(0, 0, -inactiveDays).Unix()
	protectedRelations := []string{"romantic_partner", "spouse", "family", "close_friend", "best_friend"}
	return r.DB.WithContext(r.Ctx).
		Model(&model.Memory{}).
		Where("last_access_at > 0 AND last_access_at < ? AND importance > 1", threshold).
		Where("relation_type NOT IN ?", protectedRelations).
		Where("tags NOT LIKE ? AND tags NOT LIKE ? AND tags NOT LIKE ?",
			"%important_person%", "%romantic%", "%family%").
		Update("importance", gorm.Expr("importance - 1")).Error
}

// DeleteExpired 删除过期记忆
func (r *Memory) DeleteExpired() error {
	return r.DB.WithContext(r.Ctx).
		Where("expire_at > 0 AND expire_at < ?", time.Now().Unix()).
		Delete(&model.Memory{}).Error
}

// SearchByKeyword 关键词搜索记忆
func (r *Memory) SearchByKeyword(wxID, chatRoomID, keyword string, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	query := r.DB.WithContext(r.Ctx).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix())
	if wxID != "" {
		query = query.Where("wx_id = ?", wxID)
	}
	if chatRoomID != "" {
		query = query.Where("(chat_room_id = '' OR chat_room_id = ?)", chatRoomID)
	}
	escaped := escapeLikeKeyword(keyword)
	err := query.
		Where("content LIKE ?", "%"+escaped+"%").
		Order("importance DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// escapeLikeKeyword 转义 LIKE 通配符，防止用户输入中的 % 和 _ 改变查询语义
func escapeLikeKeyword(keyword string) string {
	keyword = strings.ReplaceAll(keyword, "\\", "\\\\")
	keyword = strings.ReplaceAll(keyword, "%", "\\%")
	keyword = strings.ReplaceAll(keyword, "_", "\\_")
	return keyword
}

func (r *Memory) scopeQuery(wxID, chatRoomID string) *gorm.DB {
	query := r.DB.WithContext(r.Ctx).Model(&model.Memory{})
	if chatRoomID == "" {
		return query.Where("wx_id = ? AND chat_room_id = ''", wxID)
	}
	if wxID == "" {
		return query.Where("wx_id = '' AND chat_room_id = ?", chatRoomID)
	}
	return query.Where(
		"((wx_id = ? AND chat_room_id = '') OR (wx_id = ? AND chat_room_id = ?) OR (wx_id = '' AND chat_room_id = ?))",
		wxID,
		wxID,
		chatRoomID,
		chatRoomID,
	)
}
