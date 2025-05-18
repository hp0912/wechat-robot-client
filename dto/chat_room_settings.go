package dto

type ChatRoomSettingsRequest struct {
	ChatRoomID string `form:"chat_room_id" json:"chat_room_id" binding:"required"`
}
