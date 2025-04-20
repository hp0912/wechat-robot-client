package robot

type QrCode struct {
	BaseResponse              any    `json:"BaseResponse"`
	BlueToothBroadCastContent any    `json:"BlueToothBroadCastContent"`
	BlueToothBroadCastUuid    string `json:"BlueToothBroadCastUuid"`
	CheckTime                 int    `json:"CheckTime"`
	ExpiredTime               int    `json:"ExpiredTime"`
	NotifyKey                 any    `json:"NotifyKey"`
	Uuid                      string `json:"Uuid"`
}

type AwakenLogin struct {
	QrCodeResponse QrCode `json:"QrCodeResponse"`
}
