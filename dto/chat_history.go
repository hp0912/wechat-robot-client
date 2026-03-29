package dto

type ChatHistoryRequest struct {
	ContactID      string `form:"contact_id" json:"contact_id" binding:"required"`
	Keyword        string `form:"keyword" json:"keyword"`
	ChatRoomMember string `form:"chat_room_member" json:"chat_room_member"`
	TimeStart      int64  `form:"time_start" json:"time_start"`
	TimeEnd        int64  `form:"time_end" json:"time_end"`
}
