package model

import "gorm.io/gorm"

// ContactType 表示联系人类型的枚举
type ContactType string

const (
	ContactTypeFriend ContactType = "friend"
	ContactTypeGroup  ContactType = "group"
)

// Contact 表示微信联系人，包括好友和群组
type Contact struct {
	ID            int64          `gorm:"primarykey" json:"id"`
	WechatID      string         `gorm:"column:wechat_id;index:deleted,unique" json:"wechat_id"` // 添加索引长度
	Owner         string         `gorm:"column:owner;" json:"owner"`                             // 联系人所有者
	Alias         string         `gorm:"column:alias" json:"alias"`                              // 微信号
	Nickname      string         `gorm:"column:nickname" json:"nickname"`
	Avatar        string         `gorm:"column:avatar" json:"avatar"`
	Type          ContactType    `gorm:"column:type" json:"type"`
	Remark        string         `gorm:"column:remark" json:"remark"`
	Pyinitial     string         `gorm:"column:pyinitial" json:"pyinitial"`           // 昵称拼音首字母大写
	QuanPin       string         `gorm:"column:quan_pin" json:"quan_pin"`             // 昵称拼音全拼小写
	Sex           int            `gorm:"column:sex" json:"sex"`                       // 性别 0：未知 1：男 2：女
	Country       string         `gorm:"column:country" json:"country"`               // 国家
	Province      string         `gorm:"column:province" json:"province"`             // 省份
	City          string         `gorm:"column:city" json:"city"`                     // 城市
	Signature     string         `gorm:"column:signature" json:"signature"`           // 个性签名
	SnsBackground string         `gorm:"column:sns_background" json:"sns_background"` // 朋友圈背景图
	CreatedAt     int64          `json:"created_at"`
	UpdatedAt     int64          `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Contact) TableName() string {
	return "contacts"
}

// IsGroup 判断联系人是否为群组
func (c *Contact) IsGroup() bool {
	return c.Type == ContactTypeGroup
}
