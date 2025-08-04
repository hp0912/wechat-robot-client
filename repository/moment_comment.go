package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"

	"gorm.io/gorm"
)

type MomentComment struct {
	Ctx context.Context
	DB  *gorm.DB
}

func NewMomentCommentRepo(ctx context.Context, db *gorm.DB) *MomentComment {
	return &MomentComment{
		Ctx: ctx,
		DB:  db,
	}
}

// IsTodayHasCommented 今天这个好友的朋友圈是否被评论过了
func (respo *MomentComment) IsTodayHasCommented(contactID string) (bool, error) {
	var comment model.MomentComment
	// 获取今天凌晨零点
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayStartTimestamp := todayStart.Unix()
	err := respo.DB.WithContext(respo.Ctx).
		Where("created_at >= ? AND wechat_id = ?", todayStartTimestamp, contactID).
		First(&comment).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (respo *MomentComment) Create(data *model.MomentComment) error {
	return respo.DB.WithContext(respo.Ctx).Create(data).Error
}

func (respo *MomentComment) Update(data *model.MomentComment) error {
	return respo.DB.WithContext(respo.Ctx).Updates(data).Error
}

func (c *MomentComment) Delete(data *model.MomentComment) error {
	return c.DB.WithContext(c.Ctx).Unscoped().Delete(data).Error
}
