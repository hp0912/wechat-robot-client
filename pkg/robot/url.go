package robot

import "fmt"

const (
	LoginGetCacheInfo       = "/Login/GetCacheInfo"
	LoginTwiceAutoAuth      = "/Login/TwiceAutoAuth"
	LoginAwaken             = "/Login/Awaken"
	LoginGetQR              = "/Login/GetQRx"
	LoginCheckQR            = "/Login/CheckQR"
	LoginHeartBeat          = "/Login/HeartBeat"
	UserGetContractProfile  = "/User/GetContractProfile"
	LoginLogout             = "/Login/LogOut"
	AutoHeartbeatStartPath  = "/AutoHeartbeatStart"
	AutoHeartbeatStatusPath = "/AutoHeartbeatStatus"
	AutoHeartbeatStopPath   = "/AutoHeartbeatStop"
	MsgSyncPath             = "/Msg/Sync"
	GetContactListPath      = "/GetContractList"
	GetContactDetailPath    = "/GetContractDetail"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "http://" + string(w)
}

func (w WechatDomain) BasePath() string {
	return fmt.Sprintf("%s%s", w.BaseHost(), "/api")
}
