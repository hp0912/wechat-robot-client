package repository

import (
	"context"
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
