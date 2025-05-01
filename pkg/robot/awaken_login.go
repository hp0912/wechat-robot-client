package robot

type QrCode struct {
	BaseResponse              any           `json:"BaseResponse"`
	BlueToothBroadCastContent BuiltinBuffer `json:"BlueToothBroadCastContent"`
	BlueToothBroadCastUuid    string        `json:"BlueToothBroadCastUuid"`
	CheckTime                 int           `json:"CheckTime"`
	ExpiredTime               int           `json:"ExpiredTime"`
	NotifyKey                 BuiltinBuffer `json:"NotifyKey"`
	Uuid                      string        `json:"Uuid"`
}
