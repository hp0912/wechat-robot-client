package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserProfile struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewUserProfileRepo(ctx context.Context, db *gorm.DB) *UserProfile {
	return &UserProfile{Ctx: ctx, DB: db}
}

// Upsert 创建或更新用户画像
func (r *UserProfile) Upsert(profile *model.UserProfile) error {
	now := time.Now().Unix()
	profile.UpdatedAt = now
	if profile.CreatedAt == 0 {
		profile.CreatedAt = now
	}
	return r.DB.WithContext(r.Ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "wx_id"}, {Name: "chat_room_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"summary", "updated_at"}),
		}).Create(profile).Error
}

// GetByScope 获取指定作用域的画像
func (r *UserProfile) GetByScope(wxID, chatRoomID string) (*model.UserProfile, error) {
	var profile model.UserProfile
	err := r.DB.WithContext(r.Ctx).
		Where("wx_id = ? AND chat_room_id = ?", wxID, chatRoomID).
		First(&profile).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &profile, err
}

// GetByWxID 获取某人的所有画像（全局 + 各群）
func (r *UserProfile) GetByWxID(wxID string) ([]*model.UserProfile, error) {
	var profiles []*model.UserProfile
	err := r.DB.WithContext(r.Ctx).
		Where("wx_id = ?", wxID).
		Order("chat_room_id ASC").
		Find(&profiles).Error
	return profiles, err
}
