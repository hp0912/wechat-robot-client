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

// GetByContactAndKey 根据联系人和 key 查找记忆（用于去重合并）
func (r *Memory) GetByContactAndKey(contactWxID, key string) (*model.Memory, error) {
	var memory model.Memory
	err := r.DB.WithContext(r.Ctx).
		Where("contact_wxid = ? AND `key` = ?", contactWxID, key).
		First(&memory).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &memory, err
}

// GetByContact 获取某个联系人的所有记忆
func (r *Memory) GetByContact(contactWxID string, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	err := r.DB.WithContext(r.Ctx).
		Where("contact_wxid = ?", contactWxID).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix()).
		Order("importance DESC, updated_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}

// GetByContactAndType 获取某个联系人特定类型的记忆
func (r *Memory) GetByContactAndType(contactWxID string, memoryType model.MemoryType, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	err := r.DB.WithContext(r.Ctx).
		Where("contact_wxid = ? AND `type` = ?", contactWxID, memoryType).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix()).
		Order("importance DESC, updated_at DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
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
		Where("last_access_at < ? AND importance > 1", threshold).
		Update("importance", gorm.Expr("importance - 1")).Error
}

// DeleteExpired 删除过期记忆
func (r *Memory) DeleteExpired() error {
	return r.DB.WithContext(r.Ctx).
		Where("expire_at > 0 AND expire_at < ?", time.Now().Unix()).
		Delete(&model.Memory{}).Error
}

// SearchByKeyword 关键词搜索记忆
func (r *Memory) SearchByKeyword(contactWxID, keyword string, limit int) ([]*model.Memory, error) {
	var memories []*model.Memory
	query := r.DB.WithContext(r.Ctx).
		Where("expire_at = 0 OR expire_at > ?", time.Now().Unix())
	if contactWxID != "" {
		query = query.Where("contact_wxid = ?", contactWxID)
	}
	err := query.
		Where("content LIKE ? OR `key` LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Order("importance DESC").
		Limit(limit).
		Find(&memories).Error
	return memories, err
}
