package dto

type ChatRoomSettingsRequest struct {
	ChatRoomID string `form:"chat_room_id" json:"chat_room_id" binding:"required"`
}

type SyncChatRoomMemberRequest struct {
	ChatRoomID string `form:"chat_room_id" json:"chat_room_id" binding:"required"`
}

type ChatRoomMemberRequest struct {
	ChatRoomID string `form:"chat_room_id" json:"chat_room_id" binding:"required"`
	Keyword    string `form:"keyword" json:"keyword"`
}

// ChatRoomSummary 群动态
type ChatRoomSummary struct {
	ChatRoomID     string
	Year           int
	Month          int
	Date           int
	Week           string
	UserTotalCount int // 当前群成员总数
	UserJoinCount  int // 昨天入群数
	UserLeaveCount int // 昨天离群数
	UserChatCount  int // 昨天聊天人数
	MessageCount   int // 昨天消息数
}
