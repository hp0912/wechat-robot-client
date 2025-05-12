package dto

type MessageCommonRequest struct {
	MessageID int64 `form:"message_id" json:"message_id" binding:"required"`
}

type SendMessageCommonRequest struct {
	ToWxid string `form:"to_wxid" json:"to_wxid" binding:"required"`
}

type SendTextMessageRequest struct {
	SendMessageCommonRequest
	Content string   `form:"content" json:"content" binding:"required"`
	At      []string `form:"at" json:"at"`
}

type SendMusicMessageRequest struct {
	SendMessageCommonRequest
	Song string `form:"song" json:"song" binding:"required"`
}
