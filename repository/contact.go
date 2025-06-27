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
	Ctx context.Context
	DB  *gorm.DB
}

func NewContactRepo(ctx context.Context, db *gorm.DB) *Contact {
	return &Contact{
		Ctx: ctx,
		DB:  db,
	}
}

func (c *Contact) GetContact(wechatID string) (*model.Contact, error) {
	var contact model.Contact
	err := c.DB.WithContext(c.Ctx).Where("wechat_id = ?", wechatID).First(&contact).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (c *Contact) DeleteByWeChatIDNotIn(wechatIDs []string) error {
	return c.DB.WithContext(c.Ctx).Where("wechat_id NOT IN (?)", wechatIDs).Delete(&model.Contact{}).Error
}

func (c *Contact) FindRecentChatRoomContacts() ([]*model.Contact, error) {
	oneDayAgo := time.Now().Add(-24 * time.Hour).Unix()
	var contacts []*model.Contact
	query := c.DB.WithContext(c.Ctx).Where("type = ? AND last_active_at >= ?", model.ContactTypeChatRoom, oneDayAgo)
	if err := query.Find(&contacts).Error; err != nil {
		return nil, err
	}
	return contacts, nil
}

func (respo *Contact) GetByWechatID(wechatID string) (*model.Contact, error) {
	var contact model.Contact
	err := respo.DB.WithContext(respo.Ctx).Where("wechat_id = ?", wechatID).First(&contact).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (c *Contact) GetContacts(req dto.ContactListRequest, pager appx.Pager) ([]*model.Contact, int64, error) {
	var contacts []*model.Contact
	var total int64
	query := c.DB.WithContext(c.Ctx).Model(&model.Contact{})
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.Keyword != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Keyword+"%").Or("remark LIKE ?", "%"+req.Keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = query.Order("last_active_at DESC").Order("id DESC")
	if err := query.Offset(pager.OffSet).Limit(pager.PageSize).Find(&contacts).Error; err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

func (c *Contact) UpdateNicknameByContactID(contactID string, nickname string) error {
	return c.DB.WithContext(c.Ctx).Model(&model.Contact{}).Where("wechat_id = ?", contactID).Update("nickname", nickname).Error
}

func (c *Contact) UpdateRemarkByContactID(contactID string, remark string) error {
	return c.DB.WithContext(c.Ctx).Model(&model.Contact{}).Where("wechat_id = ?", contactID).Update("remark", remark).Error
}

func (c *Contact) Create(data *model.Contact) error {
	return c.DB.WithContext(c.Ctx).Create(data).Error
}

func (c *Contact) Update(data *model.Contact) error {
	return c.DB.WithContext(c.Ctx).Where("id = ?", data.ID).Updates(data).Error
}

func (c *Contact) DeleteByContactID(contactID string) error {
	return c.DB.WithContext(c.Ctx).Unscoped().Where("wechat_id = ?", contactID).Delete(&model.Contact{}).Error
}
