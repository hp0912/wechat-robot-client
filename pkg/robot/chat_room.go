package robot

type ChatRoomMember struct {
	BigHeadImgUrl      string  `json:"BigHeadImgUrl"`
	ChatroomMemberFlag int     `json:"ChatroomMemberFlag"`
	DisplayName        *string `json:"DisplayName"`
	InviterUserName    string  `json:"InviterUserName"`
	NickName           string  `json:"NickName"`
	SmallHeadImgUrl    string  `json:"SmallHeadImgUrl"`
	UserName           string  `json:"UserName"`
}

type NewChatroomData struct {
	ChatRoomMember []ChatRoomMember `json:"ChatRoomMember"`
	InfoMask       int              `json:"InfoMask"`
	MemberCount    int              `json:"MemberCount"`
}

type ChatRoomMemberDetail struct {
	BaseResponse
	ChatroomUserName string          `json:"ChatroomUserName"`
	ServerVersion    int64           `json:"ServerVersion"`
	NewChatroomData  NewChatroomData `json:"NewChatroomData"`
}

type GetChatRoomMemberDetailRequest struct {
	Wxid string `json:"Wxid"`
	QID  string `json:"QID"`
}
