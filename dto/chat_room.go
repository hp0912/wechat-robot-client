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
	ChatRoomID       string
	Year             int
	Month            int
	Date             int
	Week             string
	MemberTotalCount int // 当前群成员总数
	MemberJoinCount  int // 昨天入群数
	MemberLeaveCount int // 昨天离群数
	MemberChatCount  int // 昨天聊天人数
	MessageCount     int // 昨天消息数
}

type ChatRoomRank struct {
	SenderWxID string `gorm:"column:sender_wxid" json:"sender_wxid"` // 微信Id
	Nickname   string `gorm:"column:nickname" json:"nickname"`       // 昵称
	Count      int64  `gorm:"column:count" json:"count"`             // 消息数
}
