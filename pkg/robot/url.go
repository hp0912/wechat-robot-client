package robot

const (
	IsRunningPath     = "/IsRunning"
	GetProfilePath    = "/GetProfile"
	GetCachedInfoPath = "/GetCachedInfo"
	AwakenLoginPath   = "/AwakenLogin"
	GetQrCodePath     = "/GetQRCode"
	CheckUuidPath     = "/CheckUuid"
	LogoutPath        = "/Logout"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "http://" + string(w)
}
