package robot

type GetQRCode struct {
	Uuid         string `json:"Uuid"`
	QRCodeURL    string `json:"QRCodeURL"`
	QRCodeBase64 string `json:"QRCodeBase64"`
	ExpiredTime  string `json:"ExpiredTime"`
}
