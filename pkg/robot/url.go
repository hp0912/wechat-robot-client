package robot

const (
	IsRunningPath           = "/IsRunning"
	GetProfilePath          = "/GetProfile"
	GetCachedInfoPath       = "/GetCachedInfo"
	AwakenLoginPath         = "/AwakenLogin"
	GetQrCodePath           = "/GetQRCode"
	CheckUuidPath           = "/CheckUuid"
	LogoutPath              = "/LogOut"
	AutoHeartbeatStartPath  = "/AutoHeartbeatStart"
	AutoHeartbeatStatusPath = "/AutoHeartbeatStatus"
	AutoHeartbeatStopPath   = "/AutoHeartbeatStop"
	HeartbeatPath           = "/Heartbeat"
	SyncPath                = "/Sync"
	GetContactListPath      = "/GetContractList"
	GetContactDetailPath    = "/GetContractDetail"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "http://" + string(w)
}
