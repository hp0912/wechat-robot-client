package dto

type ContactListRequest struct {
	ContactIDs []string `form:"contact_ids" json:"contact_ids"`
	Type       string   `form:"type" json:"type"`
	Keyword    string   `form:"keyword" json:"keyword"`
}

type FriendSearchRequest struct {
	ToUserName  string `form:"to_username" json:"to_username" binding:"required"`
	FromScene   int    `form:"from_scene" json:"from_scene"`
	SearchScene int    `form:"search_scene" json:"search_scene"`
}

type FriendSendRequestRequest struct {
	V1            string `form:"v1" json:"V1" binding:"required"`
	V2            string `form:"v2" json:"V2" binding:"required"`
	Opcode        int    `form:"opcode" json:"Opcode"`
	Scene         int    `form:"scene" json:"Scene"`
	VerifyContent string `form:"verify_content" json:"VerifyContent"`
}

type FriendSendRequestFromChatRoomRequest struct {
	ChatRoomMemberID string `form:"chat_room_member_id" json:"chat_room_member_id" binding:"required"`
	VerifyContent    string `form:"verify_content" json:"VerifyContent"`
}

type FriendSetRemarksRequest struct {
	ToWxid  string `form:"to_wxid" json:"to_wxid" binding:"required"`
	Remarks string `form:"remarks" json:"remarks" binding:"required"`
}

type FriendPassVerifyRequest struct {
	SystemMessageID int64 `form:"system_message_id" json:"system_message_id"`
}

type FriendDeleteRequest struct {
	ContactID string `form:"contact_id" json:"contact_id"`
}
