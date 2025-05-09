package dto

type MessageCommonRequest struct {
	MessageID int64 `form:"message_id" json:"message_id" binding:"required"`
}
