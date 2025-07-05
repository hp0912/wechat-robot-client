package model

type SystemMessageType int

const (
	SystemMessageTypeVerify       SystemMessageType = 37 // 认证消息 好友请求
	SystemMessageTypeJoinChatRoom SystemMessageType = 38 // 认证消息 加入群聊
)

type SystemMessage struct {
	ID          int64             `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	MsgID       int64             `gorm:"uniqueIndex:uniq_msg_id;not null;column:msg_id" json:"msg_id"`
	ClientMsgID int64             `gorm:"not null;column:client_msg_id" json:"client_msg_id"`
	Type        SystemMessageType `gorm:"index:idx_type;not null;column:type" json:"type"`
	ImageURL    string            `gorm:"type:varchar(512);column:image_url;comment:图片URL" json:"image_url"`
	Description string            `gorm:"type:varchar(255);column:description;comment:备注" json:"description"`
	Content     string            `gorm:"type:text;column:content" json:"content"`
	FromWxid    string            `gorm:"type:varchar(64);index:idx_from_wxid;column:from_wxid" json:"from_wxid"`
	ToWxid      string            `gorm:"type:varchar(64);column:to_wxid" json:"to_wxid"`
	Status      int               `gorm:"not null;column:status;default:0;comment:'消息状态 0:未处理 1:已处理'" json:"status"`
	IsRead      bool              `gorm:"column:is_read;default:false;comment:'消息是否已读'" json:"is_read"`
	CreatedAt   int64             `gorm:"index:idx_created_at;not null;column:created_at" json:"created_at"`
	UpdatedAt   int64             `gorm:"not null;column:updated_at" json:"updated_at"`
}

func (SystemMessage) TableName() string {
	return "system_messages"
}
