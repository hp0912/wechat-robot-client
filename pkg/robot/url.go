package robot

const (
	IsRunning  = "/IsRunning"
	GetProfile = "/GetProfile"
)

type WechatDomain string

func (w WechatDomain) BaseHost() string {
	return "http://" + string(w)
}
