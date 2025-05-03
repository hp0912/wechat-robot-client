package model

import (
	"wechat-robot-client/pkg/robot"

	"gorm.io/gorm"
)

type Message struct {
	ID                 int64             `gorm:"primarykey" json:"id"`
	MsgId              int64             `gorm:"column:msg_id;index;" json:"msg_id"`               // 消息Id
	ClientMsgId        int64             `gorm:"column:client_msg_id;index;" json:"client_msg_id"` // 客户端消息Id
	IsGroup            bool              `gorm:"column:is_group;default:false;comment:'消息是否来自群聊'" json:"is_group"`
	IsAtMe             bool              `gorm:"column:is_atme;default:false;comment:'消息是否艾特我'" json:"is_atme"` // @所有人 好的
	IsRecalled         bool              `gorm:"column:is_recalled;default:false;comment:'消息是否已经撤回'" json:"is_recalled"`
	Type               robot.MessageType `gorm:"column:type" json:"type"`                                 // 消息类型
	Content            string            `gorm:"column:content" json:"content"`                           // 内容
	DisplayFullContent string            `gorm:"column:display_full_content" json:"display_full_content"` // 显示的完整内容
	MessageSource      string            `gorm:"column:message_source" json:"message_source"`
	FromWxID           string            `gorm:"column:from_wxid" json:"from_wxid"`           // 消息来源
	SenderWxID         string            `gorm:"column:sender_wxid" json:"sender_wxid"`       // 消息发送者
	ToWxID             string            `gorm:"column:to_wxid" json:"to_wxid"`               // 接收者
	AttachmentUrl      string            `gorm:"column:attachment_url" json:"attachment_url"` // 文件地址
	CreatedAt          int64             `json:"created_at"`
	UpdatedAt          int64             `json:"updated_at"`
	DeletedAt          gorm.DeletedAt    `json:"-" gorm:"index"`
	// 额外字段，通过联表查询填充，不参与建表
	SenderNickname string `gorm:"column:sender_nickname;-:migration" json:"sender_nickname"`
	SenderAvatar   string `gorm:"column:sender_avatar;-:migration" json:"sender_avatar"`
}

// TableName 指定表名
func (Message) TableName() string {
	return "messages"
}
