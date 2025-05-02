package dto

type SyncChatRoomMemberRequest struct {
	ChatRoomID string `form:"chat_room_id" json:"chat_room_id" binding:"required"`
}

type ChatRoomMemberRequest struct {
	ChatRoomID string `form:"chat_room_id" json:"chat_room_id" binding:"required"`
	Keyword    string `form:"keyword" json:"keyword"`
}
