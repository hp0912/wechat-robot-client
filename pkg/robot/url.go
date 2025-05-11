package robot

import "fmt"

const (
	LoginGetCacheInfo  = "/Login/GetCacheInfo"
	LoginTwiceAutoAuth = "/Login/TwiceAutoAuth"
	LoginAwaken        = "/Login/Awaken"
	LoginGetQR         = "/Login/GetQRx"
	LoginCheckQR       = "/Login/CheckQR"
	LoginHeartBeat     = "/Login/HeartBeat"
	LoginLogout        = "/Login/LogOut"

	UserGetContactProfile = "/User/GetContractProfile"

	MsgSyncPath  = "/Msg/Sync"
	MsgRevoke    = "/Msg/Revoke"
	MsgSendTxt   = "/Msg/SendTxt"
	MsgUploadImg = "/Msg/UploadImg"
	MsgSendVideo = "/Msg/SendVideo"
	MsgSendVoice = "/Msg/SendVoice"

	FriendGetContactList         = "/Friend/GetContractList"
	FriendGetContactDetail       = "/Friend/GetContractDetail"
	GroupGetChatRoomMemberDetail = "/Group/GetChatRoomMemberDetail"

	ToolsCdnDownloadImage = "/Tools/CdnDownloadImage"
	ToolsDownloadVideo    = "/Tools/DownloadVideo"
	ToolsDownloadVoice    = "/Tools/DownloadVoice"
	ToolsDownloadFile     = "/Tools/DownloadFile"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "http://" + string(w)
}

func (w WechatDomain) BasePath() string {
	return fmt.Sprintf("%s%s", w.BaseHost(), "/api")
}
