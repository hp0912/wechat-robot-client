package robot

import "fmt"

const (
	LoginGetCacheInfo         = "/Login/GetCacheInfo"
	LoginTwiceAutoAuth        = "/Login/LoginTwiceAutoAuth"
	LoginAwaken               = "/Login/LoginAwaken"
	LoginGetQR                = "/Login/LoginGetQR"
	LoginCheckQR              = "/Login/LoginCheckQR"
	LoginYPayVerificationcode = "/Login/YPayVerificationcode"
	AutoHeartBeat             = "/Login/AutoHeartBeat"
	CloseAutoHeartBeat        = "/Login/CloseAutoHeartBeat"
	LoginHeartBeat            = "/Login/HeartBeat"
	LoginLogout               = "/Login/LogOut"

	UserGetContactProfile = "/User/GetContractProfile"

	MsgSync         = "/Msg/Sync"
	MsgRevoke       = "/Msg/Revoke"
	MsgSendTxt      = "/Msg/SendTxt"
	MsgUploadImg    = "/Msg/UploadImg"
	MsgSendVideo    = "/Msg/SendVideo"
	MsgSendVoice    = "/Msg/SendVoice"
	MsgSendApp      = "/Msg/SendApp"
	MsgSendEmoji    = "/Msg/SendEmoji"
	MsgShareLink    = "/Msg/ShareLink"
	MsgSendCDNFile  = "/Msg/SendCDNFile"
	MsgSendCDNImg   = "/Msg/SendCDNImg"
	MsgSendCDNVideo = "/Msg/SendCDNVideo"

	FriendGetFriendstate   = "/Friend/GetFriendstate"
	FriendSearch           = "/Friend/Search"
	FriendSendRequest      = "/Friend/SendRequest"
	FriendSetRemarks       = "/Friend/SetRemarks"
	FriendGetContactList   = "/Friend/GetContractList"
	FriendGetContactDetail = "/Friend/GetContractDetail"
	FriendPassVerify       = "/Friend/PassVerify"
	FriendDelete           = "/Friend/Delete"

	GroupCreateChatRoom          = "/Group/CreateChatRoom"
	GroupAddChatRoomMember       = "/Group/AddChatRoomMember"
	GroupInviteChatRoomMember    = "/Group/InviteChatRoomMember"
	GroupConsentToJoin           = "/Group/ConsentToJoin"
	GroupGetChatRoomMemberDetail = "/Group/GetChatRoomMemberDetail"
	GroupSetChatRoomName         = "/Group/SetChatRoomName"
	GroupSetChatRoomRemarks      = "/Group/SetChatRoomRemarks"
	GroupSetChatRoomAnnouncement = "/Group/SetChatRoomAnnouncement"
	GroupDelChatRoomMember       = "/Group/DelChatRoomMember"
	GroupQuit                    = "/Group/Quit"

	FriendCircleComment               = "/FriendCircle/Comment"
	FriendCircleGetDetail             = "/FriendCircle/GetDetail"
	FriendCircleGetIdDetail           = "/FriendCircle/GetIdDetail"
	FriendCircleDownFriendCircleMedia = "/FriendCircle/DownFriendCircleMedia"
	FriendCircleGetList               = "/FriendCircle/GetList"
	FriendCircleMessages              = "/FriendCircle/Messages"
	FriendCircleMmSnsSync             = "/FriendCircle/MmSnsSync"
	FriendCircleOperation             = "/FriendCircle/Operation"
	FriendCirclePrivacySettings       = "/FriendCircle/PrivacySettings"
	FriendCircleUpload                = "/FriendCircle/Upload"
	FriendCircleCdnSnsUploadVideo     = "/FriendCircle/CdnSnsUploadVideo"

	ToolsCdnDownloadImage = "/Tools/CdnDownloadImage"
	ToolsDownloadVideo    = "/Tools/DownloadVideo"
	ToolsDownloadVoice    = "/Tools/DownloadVoice"
	ToolsDownloadFile     = "/Tools/DownloadFile"

	WxappQrcodeAuthLogin = "/Wxapp/QrcodeAuthLogin"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "http://" + string(w)
}

func (w WechatDomain) BasePath() string {
	return fmt.Sprintf("%s%s", w.BaseHost(), "/api")
}
