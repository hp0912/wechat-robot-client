package robot

const (
	IsRunningPath       = "/IsRunning"
	GetProfilePath      = "/GetProfile"
	GetCachedInfoPath   = "/GetCachedInfo"
	AwakenLoginPath     = "/AwakenLogin"
	GetQrCodePath       = "/GetQRCode"
	CheckUuidPath       = "/CheckUuid"
	LogoutPath          = "/LogOut"
	AutoHeartbeatStart  = "/AutoHeartbeatStart"
	AutoHeartbeatStatus = "/AutoHeartbeatStatus"
	AutoHeartbeatStop   = "/AutoHeartbeatStop"
	Heartbeat           = "/Heartbeat"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "http://" + string(w)
}
