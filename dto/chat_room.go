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

type ChatRoomRequestBase struct {
	ChatRoomID string `form:"chat_room_id" json:"chat_room_id" binding:"required"`
}

type ChatRoomOperateRequest struct {
	ChatRoomRequestBase
	Content string `form:"content" json:"content" binding:"required"`
}

type DelChatRoomMemberRequest struct {
	ChatRoomRequestBase
	MemberIDs []string `form:"member_ids" json:"member_ids" binding:"required"`
}

type GroupConsentToJoinRequest struct {
	SystemMessageID int64 `form:"system_message_id" json:"system_message_id"`
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
	SenderWxID             string `gorm:"column:sender_wxid" json:"sender_wxid"`                             // 微信Id
	ChatRoomMemberNickname string `gorm:"column:chat_room_member_nickname" json:"chat_room_member_nickname"` // 昵称
	Count                  int64  `gorm:"column:count" json:"count"`                                         // 消息数
}
