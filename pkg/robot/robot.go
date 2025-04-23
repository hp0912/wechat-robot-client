package robot

import (
	"context"
	"errors"
	"wechat-robot-client/model"
)

type Robot struct {
	RobotID          int64
	WxID             string
	Status           model.RobotStatus
	DeviceID         string
	DeviceName       string
	Client           *Client
	HeartbeatContext context.Context
	HeartbeatCancel  func()
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

func (r *Robot) GetQrCode() (uuid string, err error) {
	var resp GetQRCode
	resp, err = r.Client.GetQrCode(r.DeviceID, r.DeviceName)
	if err != nil {
		return
	}
	if resp.Uuid != "" && resp.QRCodeURL != "" {
		uuid = resp.Uuid
		return
	}
	err = errors.New("获取二维码失败")
	return
}

func (r *Robot) GetProfile(wxid string) (UserProfile, error) {
	return r.Client.GetProfile(wxid)
}

func (r *Robot) Login() (uuid string, awken bool, err error) {
	// 尝试唤醒登陆
	var cachedInfo CachedInfo
	cachedInfo, err = r.Client.GetCachedInfo(r.WxID)
	if err == nil && cachedInfo.Wxid != "" {
		// 唤醒登陆
		var resp AwakenLogin
		resp, err = r.Client.AwakenLogin(r.WxID)
		if err != nil {
			// 如果唤醒失败，尝试获取二维码
			uuid, err = r.GetQrCode()
			return
		}
		if resp.QrCodeResponse.Uuid == "" {
			// 如果唤醒失败，尝试获取二维码
			uuid, err = r.GetQrCode()
			return
		}
		// 唤醒登陆成功
		uuid = resp.QrCodeResponse.Uuid
		awken = true
		return
	}
	// 二维码登陆
	uuid, err = r.GetQrCode()
	return
}

func (r *Robot) CheckLoginUuid(uuid string) (CheckUuid, error) {
	return r.Client.CheckLoginUuid(uuid)
}

func (r *Robot) Logout() error {
	if r.WxID == "" {
		return errors.New("您还未登陆")
	}
	return r.Client.Logout(r.WxID)
}

func (r *Robot) AutoHeartbeatStop() error {
	return r.Client.AutoHeartbeatStop(r.WxID)
}

func (r *Robot) AutoHeartbeatStart() error {
	return r.Client.AutoHeartbeatStart(r.WxID)
}

func (r *Robot) Heartbeat() error {
	return r.Client.Heartbeat(r.WxID)
}

func (r *Robot) AutoHeartbeatStatus() (bool, error) {
	return r.Client.AutoHeartbeatStatus(r.WxID)
}
