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

type ConsentToJoinRequest struct {
	Wxid string `json:"Wxid"`
	Url  string `json:"Url"`
}

type InviteChatRoomMemberRequest struct {
	Wxid         string `json:"Wxid"`
	ChatRoomName string `json:"ChatRoomName"`
	ToWxids      string `json:"ToWxids"`
}

type InviteChatRoomMemberResponse struct {
	BaseResponse *BaseResponse `json:"baseResponse,omitempty"`
	MemberCount  *uint32       `json:"MemberCount,omitempty"`
	MemberList   []*MemberResp `json:"MemberList,omitempty"`
}

type MemberResp struct {
	MemberName      *SKBuiltinStringT `json:"MemberName,omitempty"`
	MemberStatus    *uint32           `json:"MemberStatus,omitempty"`
	NickName        *SKBuiltinStringT `json:"NickName,omitempty"`
	PYInitial       *SKBuiltinStringT `json:"PYInitial,omitempty"`
	QuanPin         *SKBuiltinStringT `json:"QuanPin,omitempty"`
	Sex             *int32            `json:"Sex,omitempty"`
	Remark          *SKBuiltinStringT `json:"Remark,omitempty"`
	RemarkPyinitial *SKBuiltinStringT `json:"RemarkPyinitial,omitempty"`
	RemarkQuanPin   *SKBuiltinStringT `json:"RemarkQuanPin,omitempty"`
	ContactType     *uint32           `json:"ContactType,omitempty"`
	Province        *string           `json:"Province,omitempty"`
	City            *string           `json:"City,omitempty"`
	Signature       *string           `json:"Signature,omitempty"`
	PersonalCard    *uint32           `json:"PersonalCard,omitempty"`
	VerifyFlag      *uint32           `json:"VerifyFlag,omitempty"`
	VerifyInfo      *string           `json:"VerifyInfo,omitempty"`
	Country         *string           `json:"Country,omitempty"`
}
