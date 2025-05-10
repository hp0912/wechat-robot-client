package dto

type MessageCommonRequest struct {
	MessageID int64 `form:"message_id" json:"message_id" binding:"required"`
}

type SendTextMessageRequest struct {
	ToWxid  string   `form:"to_wxid" json:"to_wxid" binding:"required"`
	Content string   `form:"content" json:"content" binding:"required"`
	At      []string `form:"at" json:"at"`
}
