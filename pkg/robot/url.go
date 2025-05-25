package robot

import "fmt"

const (
	LoginGetCacheInfo  = "/Login/GetCacheInfo"
	LoginTwiceAutoAuth = "/Login/LoginTwiceAutoAuth"
	LoginAwaken        = "/Login/LoginAwaken"
	LoginGetQR         = "/Login/LoginGetQRx"
	LoginCheckQR       = "/Login/LoginCheckQR"
	LoginHeartBeat     = "/Login/HeartBeat"
	LoginLogout        = "/Login/LogOut"

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
