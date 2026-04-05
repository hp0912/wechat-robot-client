package repository

import (
	"context"
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
	err := r.DB.WithContext(r.Ctx).Where("id IN ?", ids).Find(&memories).Error
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
func (r *Memory) DecayMemories(inactiveDays int) error {
	threshold := time.Now().AddDate(0, 0, -inactiveDays).Unix()
	return r.DB.WithContext(r.Ctx).
		Model(&model.Memory{}).
		Where("last_access_at > 0 AND last_access_at < ? AND importance > 1", threshold).
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
	err := query.
		Where("content LIKE ?", "%"+keyword+"%").
		Order("importance DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}
