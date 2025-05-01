package repository

import (
	"context"
	"time"
	"wechat-robot-client/model"

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

func (c *Contact) FindByOwner(owner string, preloads ...string) []*model.Contact {
	var contacts []*model.Contact
	query := c.DB.Where("owner = ?", owner).Order("updated_at DESC")
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	c.panicError(query.Find(&contacts).Error)
	return contacts
}
