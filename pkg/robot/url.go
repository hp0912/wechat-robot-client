package robot

import "fmt"

const (
	LoginGetCacheInfo            = "/Login/GetCacheInfo"
	LoginTwiceAutoAuth           = "/Login/TwiceAutoAuth"
	LoginAwaken                  = "/Login/Awaken"
	LoginGetQR                   = "/Login/GetQRx"
	LoginCheckQR                 = "/Login/CheckQR"
	LoginHeartBeat               = "/Login/HeartBeat"
	LoginLogout                  = "/Login/LogOut"
	UserGetContactProfile        = "/User/GetContractProfile"
	MsgSyncPath                  = "/Msg/Sync"
	FriendGetContactList         = "/Friend/GetContractList"
	FriendGetContactDetail       = "/Friend/GetContractDetail"
	GroupGetChatRoomMemberDetail = "/Group/GetChatRoomMemberDetail"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "http://" + string(w)
}

func (w WechatDomain) BasePath() string {
	return fmt.Sprintf("%s%s", w.BaseHost(), "/api")
}
