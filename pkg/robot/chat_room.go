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

type ChatRoomRequestBase struct {
	Wxid string `json:"Wxid"`
	QID  string `json:"QID"`
}

type OperateChatRoomInfoParam struct {
	Wxid    string
	QID     string
	Content string
}

type DelChatRoomMemberRequest struct {
	Wxid         string `json:"Wxid"`
	ChatRoomName string `json:"ChatRoomName"`
	ToWxids      string `json:"ToWxids"`
}

type DelMemberResp struct {
	MemberName *SKBuiltinStringT `json:"MemberName,omitempty"`
}

type DelChatRoomMemberResponse struct {
	BaseResponse *BaseResponse    `json:"baseResponse,omitempty"`
	MemberCount  *uint32          `json:"MemberCount,omitempty"`
	MemberList   []*DelMemberResp `json:"MemberList,omitempty"`
}
