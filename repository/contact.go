package repository

import (
	"context"
	"time"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"

	"gorm.io/gorm"
)

type Contact struct {
	Base[model.Contact]
}

func NewContactRepo(ctx context.Context, db *gorm.DB) *Contact {
	return &Contact{
		Base[model.Contact]{
			Ctx: ctx,
			DB:  db,
		}}
}

func (c *Contact) ExistsByWeChatID(wechatID string) bool {
	return c.ExistsByWhere(where{
		"wechat_id": wechatID,
	})
}

func (c *Contact) DeleteByWeChatIDNotIn(wechatIDs []string) {
	c.panicError(c.DB.Where("wechat_id NOT IN (?)", wechatIDs).Delete(new(model.Contact)).Error)
}

func (c *Contact) FindRecentGroupContacts(preloads ...string) []*model.Contact {
	oneDayAgo := time.Now().Add(-24 * time.Hour).Unix()
	var contacts []*model.Contact
	query := c.DB.Where("type = ? AND updated_at >= ?", model.ContactTypeGroup, oneDayAgo)
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	c.panicError(query.Find(&contacts).Error)
	return contacts
}

func (c *Contact) FindByOwner(req dto.ContactListRequest, pager appx.Pager, preloads ...string) ([]*model.Contact, int64, error) {
	var contacts []*model.Contact
	var total int64

	query := c.DB.Model(&model.Contact{})
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Keyword != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Keyword+"%").
			Or("alias LIKE ?", "%"+req.Keyword+"%").
			Or("wechat_id LIKE ?", "%"+req.Keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("updated_at DESC").Order("id DESC")
	if err := query.Offset(pager.OffSet).Limit(pager.PageSize).Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}
