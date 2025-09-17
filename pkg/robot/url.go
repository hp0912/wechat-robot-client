package robot

import "fmt"

const (
	LoginGetCacheInfo         = "/Login/GetCacheInfo"
	LoginTwiceAutoAuth        = "/Login/LoginTwiceAutoAuth"
	LoginAwaken               = "/Login/LoginAwaken"
	LoginGetQR                = "/Login/LoginGetQR"
	LoginGetQRMac             = "/Login/LoginGetQRMac"
	LoginCheckQR              = "/Login/LoginCheckQR"
	LoginYPayVerificationcode = "/Login/YPayVerificationcode"
	LoginAutoHeartBeat        = "/Login/AutoHeartBeat"
	LoginCloseAutoHeartBeat   = "/Login/CloseAutoHeartBeat"
	LoginHeartBeat            = "/Login/HeartBeat"
	LoginNewDeviceVerify      = "/Login/NewDeviceVerify"
	LoginGet62Data            = "/Login/Get62Data"
	LoginGetA16Data           = "/Login/GetA16Data"
	LoginA16Data1             = "/Login/A16Data"
	LoginData62SMSApply       = "/Login/Data62SMSApply"
	LoginData62SMSAgain       = "/Login/Data62SMSAgain"
	LoginData62SMSVerify      = "/Login/Data62SMSVerify"
	LoginLogout               = "/Login/LogOut"

	UserGetContactProfile = "/User/GetContractProfile"

	MsgSync                 = "/Msg/Sync"
	MsgRevoke               = "/Msg/Revoke"
	MsgSendTxt              = "/Msg/SendTxt"
	MsgUploadImg            = "/Msg/UploadImg"
	MsgSendVideo            = "/Msg/SendVideo"
	MsgSendVoice            = "/Msg/SendVoice"
	MsgSendApp              = "/Msg/SendApp"
	MsgSendEmoji            = "/Msg/SendEmoji"
	MsgShareLink            = "/Msg/ShareLink"
	MsgSendCDNFile          = "/Msg/SendCDNFile"
	MsgSendCDNImg           = "/Msg/SendCDNImg"
	MsgSendCDNVideo         = "/Msg/SendCDNVideo"
	MsgSendGroupMassMsgText = "/Msg/SendGroupMassMsgText" // 文本消息，群发接口

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

	ToolsCdnDownloadImage      = "/Tools/CdnDownloadImage"
	ToolsDownloadVideo         = "/Tools/DownloadVideo"
	ToolsDownloadVoice         = "/Tools/DownloadVoice"
	ToolsDownloadFile          = "/Tools/DownloadFile"
	ToolsUploadAppAttachStream = "/Tools/UploadAppAttachStream"

	WxappQrcodeAuthLogin = "/Wxapp/QrcodeAuthLogin"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "http://" + string(w)
}

func (w WechatDomain) BasePath() string {
	return fmt.Sprintf("%s%s", w.BaseHost(), "/api")
}
