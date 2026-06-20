package model

import "gorm.io/gorm"

// ContactType 表示联系人类型的枚举
type ContactType string

const (
	ContactTypeFriend          ContactType = "friend"
	ContactTypeChatRoom        ContactType = "chat_room"
	ContactTypeOfficialAccount ContactType = "official_account"
)

// Contact 表示微信联系人，包括好友和群组
type Contact struct {
	ID            int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement;not null;comment:主键ID" json:"id"`
	WechatID      string         `gorm:"column:wechat_id;type:varchar(64);not null;uniqueIndex:uk_wechat_id;comment:微信ID" json:"wechat_id"` // 微信号
	Alias         string         `gorm:"column:alias;type:varchar(64);comment:微信号" json:"alias"`                                            // 微信号别名
	Nickname      *string        `gorm:"column:nickname;type:varchar(64);comment:昵称" json:"nickname"`
	Avatar        string         `gorm:"column:avatar;type:varchar(255);comment:头像" json:"avatar"`
	Type          ContactType    `gorm:"column:type;type:enum('friend','chat_room','official_account');not null;comment:联系人类型：friend-好友，chat_room-群组，official_account-公众号" json:"type"`
	Status        *int           `gorm:"column:status;type:int;not null;default:0;comment:状态" json:"status"`
	Remark        string         `gorm:"column:remark;type:varchar(255);comment:备注" json:"remark"`
	Pyinitial     *string        `gorm:"column:pyinitial;type:varchar(64);comment:昵称拼音首字母大写" json:"pyinitial"`                                 // 昵称拼音首字母大写
	QuanPin       *string        `gorm:"column:quan_pin;type:varchar(255);comment:昵称拼音全拼小写" json:"quan_pin"`                                   // 昵称拼音全拼小写
	Sex           int            `gorm:"column:sex;type:tinyint;default:0;comment:性别 0：未知 1：男 2：女" json:"sex"`                                 // 性别 0：未知 1：男 2：女
	Country       string         `gorm:"column:country;type:varchar(64);comment:国家" json:"country"`                                            // 国家
	Province      string         `gorm:"column:province;type:varchar(64);comment:省份" json:"province"`                                          // 省份
	City          string         `gorm:"column:city;type:varchar(64);comment:城市" json:"city"`                                                  // 城市
	Signature     string         `gorm:"column:signature;type:varchar(255);comment:个性签名" json:"signature"`                                     // 个性签名
	SnsBackground *string        `gorm:"column:sns_background;type:varchar(255);comment:朋友圈背景图" json:"sns_background"`                         // 朋友圈背景图
	ChatRoomOwner string         `gorm:"column:chat_room_owner;type:varchar(64);not null;default:'';comment:群聊所有者微信ID" json:"chat_room_owner"` // 群主微信号
	CreatedAt     int64          `gorm:"column:created_at;type:bigint;not null;comment:创建时间" json:"created_at"`
	LastActiveAt  int64          `gorm:"column:last_active_at;type:bigint;not null;index:idx_last_active_at;comment:最近活跃时间" json:"last_active_at"` // 最近活跃时间
	UpdatedAt     int64          `gorm:"column:updated_at;type:bigint;not null;comment:更新时间" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:软删除时间" json:"-"`
}

// TableName 指定表名
func (Contact) TableName() string {
	return "contacts"
}

// IsChatRoom 判断联系人是否为群组
func (c *Contact) IsChatRoom() bool {
	return c.Type == ContactTypeChatRoom
}
