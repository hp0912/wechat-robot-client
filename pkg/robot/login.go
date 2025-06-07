package robot

type LoginGetQRRequest struct {
	DeviceID   string    `json:"DeviceID"`
	DeviceName string    `json:"DeviceName"`
	LoginType  string    `json:"LoginType"`
	Proxy      ProxyInfo `json:"Proxy"`
}

type GetQRCode struct {
	Uuid         string `json:"Uuid"`
	QRCodeURL    string `json:"QRCodeURL"`
	QRCodeBase64 string `json:"QRCodeBase64"`
	ExpiredTime  string `json:"ExpiredTime"`
}

type AwakenLoginRequest struct {
	Wxid  string    `json:"Wxid"`
	Proxy ProxyInfo `json:"Proxy"`
}

type UnifyAuthResponse struct {
	BaseResponse      *BaseResponse `json:"baseResponse,omitempty"`
	UnifyAuthSectFlag *uint32       `json:"unifyAuthSectFlag,omitempty"`
}

type QrCode struct {
	BaseResponse              BaseResponse     `json:"BaseResponse"`
	BlueToothBroadCastContent SKBuiltinBufferT `json:"BlueToothBroadCastContent"`
	BlueToothBroadCastUuid    string           `json:"BlueToothBroadCastUuid"`
	CheckTime                 int              `json:"CheckTime"`
	ExpiredTime               int              `json:"ExpiredTime"`
	NotifyKey                 SKBuiltinBufferT `json:"NotifyKey"`
	Uuid                      string           `json:"Uuid"`
}

type AcctSectRespData struct {
	Username   string `json:"userName"`   // 原始微信Id
	Alias      string `json:"alias"`      // 自定义的微信号
	BindMobile string `json:"bindMobile"` // 绑定的手机号
	FsUrl      string `json:"fsurl"`      // 可能是头像地址
	Nickname   string `json:"nickName"`   // 昵称
}

type CheckUuid struct {
	Uuid                    string           `json:"uuid"`
	Status                  int              `json:"status"` // 状态
	PushLoginUrlexpiredTime int              `json:"pushLoginUrlexpiredTime"`
	ExpiredTime             int              `json:"expiredTime"`  // 过期时间(秒)
	HeadImgUrl              string           `json:"headImgUrl"`   // 头像
	NickName                string           `json:"nickName"`     // 昵称
	AcctSectResp            AcctSectRespData `json:"acctSectResp"` // 账号信息-登录成功之后才有
}
