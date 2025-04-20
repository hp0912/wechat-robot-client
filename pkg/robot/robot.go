package robot

import "errors"

type Robot struct {
	RobotID    int64
	WxID       string
	DeviceID   string
	DeviceName string
	Client     *Client
}

func (r *Robot) IsRunning() bool {
	return r.Client.IsRunning()
}

func (r *Robot) IsLoggedIn() bool {
	if r.WxID == "" {
		return false
	}
	_, err := r.Client.GetProfile(r.WxID)
	return err == nil
}

func (r *Robot) GetQrCode() (uuid, qrcode string, err error) {
	var resp GetQRCode
	resp, err = r.Client.GetQrCode(r.DeviceID, r.DeviceName)
	if err != nil {
		return
	}
	if resp.Uuid != "" && resp.QRCodeURL != "" {
		uuid = resp.Uuid
		qrcode = resp.QRCodeURL
		return
	}
	err = errors.New("获取二维码失败")
	return
}

func (r *Robot) Login() (profile UserProfile, uuid, qrcode string, err error) {
	if r.IsLoggedIn() {
		profile, err = r.Client.GetProfile(r.WxID)
		return
	}
	// 尝试唤醒登陆
	var cachedInfo CachedInfo
	cachedInfo, err = r.Client.GetCachedInfo(r.WxID)
	if err == nil && cachedInfo.Wxid != "" {
		// 唤醒登陆
		var resp AwakenLogin
		resp, err = r.Client.AwakenLogin(r.WxID)
		if err != nil {
			// 如果唤醒失败，尝试获取二维码
			uuid, qrcode, err = r.GetQrCode()
			return
		}
		if resp.QrCodeResponse.Uuid == "" {
			// 如果唤醒失败，尝试获取二维码
			uuid, qrcode, err = r.GetQrCode()
			return
		}
		// 唤醒登陆成功
		profile, err = r.Client.GetProfile(r.WxID)
		return
	}
	// 二维码登陆
	uuid, qrcode, err = r.GetQrCode()
	return
}

func (r *Robot) CheckLoginUuid() (CheckUuid, error) {
	return r.Client.CheckLoginUuid(r.WxID)
}

func (r *Robot) Logout() error {
	if r.WxID == "" {
		return errors.New("您还未登陆")
	}
	return r.Client.Logout(r.WxID)
}
