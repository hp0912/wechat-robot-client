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
	Data62       string `json:"Data62"`
	ExpiredTime  string `json:"ExpiredTime"`
}

type AwakenLoginRequest struct {
	Wxid  string    `json:"Wxid"`
	Proxy ProxyInfo `json:"Proxy"`
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
	Ticket                  string           `json:"ticket"`       // 登录票据
	AcctSectResp            AcctSectRespData `json:"acctSectResp"` // 账号信息-登录成功之后才有
}

type LoginResponse struct {
	Uuid       string `json:"uuid"`
	Data62     string `json:"data62"`
	AwkenLogin bool   `json:"awken_login"`
	AutoLogin  bool   `json:"auto_login"`
}

type VerificationCodeRequest struct {
	Uuid   string
	Data62 string
	Code   string
	Ticket string
}

type SilderOCR struct {
	Flag    int    `json:"flag"`
	Data    string `json:"data"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	URL     string `json:"url"`
	Remark  string `json:"remark"`
	Success bool   `json:"Success"`
}

type A16LoginRequest struct {
	UserName   string
	Password   string
	A16        string
	DeviceName string
	Proxy      ProxyInfo
}

type Data62LoginRequest struct {
	UserName   string
	Password   string
	Data62     string
	DeviceName string
	Proxy      ProxyInfo
}

type UnifyAuthResponse struct {
	BaseResponse      *BaseResponse `json:"baseResponse,omitempty"`
	UnifyAuthSectFlag *uint32       `json:"unifyAuthSectFlag,omitempty"`
	AuthSectResp      *AuthSectResp `json:"authSectResp,omitempty"`
	// AcctSectResp      *AcctSectResp      `json:"acctSectResp,omitempty"`
	// NetworkSectResp   *NetworkSectResp   `json:"networkSectResp,omitempty"`
	// AxAuthSecRespList *AxAuthSecRespList `json:"axAuthSecRespList,omitempty"`
}

type AuthSectResp struct {
	Uin                  *uint32               `json:"uin,omitempty"`
	SvrPubEcdhkey        *ECDHKey              `json:"svrPubEcdhkey,omitempty"`
	SessionKey           *SKBuiltinBufferT     `json:"sessionKey,omitempty"`
	AutoAuthKey          *SKBuiltinBufferT     `json:"autoAuthKey,omitempty"`
	WtloginRspBuffFlag   *uint32               `json:"wtloginRspBuffFlag,omitempty"`
	WtloginRspBuff       *SKBuiltinBufferT     `json:"wtloginRspBuff,omitempty"`
	WtloginImgRespInfo   *WTLoginImgRespInfo   `json:"wtloginImgRespInfo,omitempty"`
	WxVerifyCodeRespInfo *WxVerifyCodeRespInfo `json:"wxVerifyCodeRespInfo,omitempty"`
	CliDbencryptKey      *SKBuiltinBufferT     `json:"cliDbencryptKey,omitempty"`
	CliDbencryptInfo     *SKBuiltinBufferT     `json:"cliDbencryptInfo,omitempty"`
	AuthKey              *string               `json:"authKey,omitempty"`
	A2Key                *SKBuiltinBufferT     `json:"a2Key,omitempty"`
	ApplyBetaUrl         *string               `json:"applyBetaUrl,omitempty"`
	ShowStyle            *ShowStyleKey         `json:"showStyle,omitempty"`
	AuthTicket           *string               `json:"authTicket,omitempty"`
	NewVersion           *uint32               `json:"newVersion,omitempty"`
	UpdateFlag           *uint32               `json:"updateFlag,omitempty"`
	AuthResultFlag       *uint32               `json:"authResultFlag,omitempty"`
	Fsurl                *string               `json:"fsurl,omitempty"`
	MmtlsControlBitFlag  *uint32               `json:"mmtlsControlBitFlag,omitempty"`
	ServerTime           *uint32               `json:"serverTime,omitempty"`
	ClientSessionKey     *SKBuiltinBufferT     `json:"clientSessionKey,omitempty"`
	ServerSessionKey     *SKBuiltinBufferT     `json:"serverSessionKey,omitempty"`
	EcdhControlFlag      *uint32               `json:"ecdhControlFlag,omitempty"`
}

type ECDHKey struct {
	Nid *int32            `json:"nid,omitempty"`
	Key *SKBuiltinBufferT `json:"key,omitempty"`
}

type WTLoginImgRespInfo struct {
	ImgEncryptKey *string           `json:"imgEncryptKey,omitempty"`
	Ksid          *SKBuiltinBufferT `json:"ksid,omitempty"`
	ImgSid        *string           `json:"imgSid,omitempty"`
	ImgBuf        *SKBuiltinBufferT `json:"imgBuf,omitempty"`
}

type ShowStyleKey struct {
	KeyCount *uint32 `json:"keyCount,omitempty"`
	// Key      []string `json:"key,omitempty"`
}

type WxVerifyCodeRespInfo struct {
	VerifySignature *string           `json:"verifySignature,omitempty"`
	VerifyBuff      *SKBuiltinBufferT `json:"verifyBuff,omitempty"`
}
