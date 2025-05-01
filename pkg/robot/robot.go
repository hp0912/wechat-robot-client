package robot

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

type Robot struct {
	RobotID            int64
	WxID               string
	Status             RobotStatus
	DeviceID           string
	DeviceName         string
	Client             *Client
	HeartbeatContext   context.Context
	HeartbeatCancel    func()
	SyncMessageContext context.Context
	SyncMessageCancel  func()
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
	if resp.Uuid != "" {
		uuid = resp.Uuid
		return
	}
	err = errors.New("获取二维码失败")
	return
}

func (r *Robot) GetProfile(wxid string) (UserProfile, error) {
	return r.Client.GetProfile(wxid)
}

func (r *Robot) Login() (uuid string, awkenLogin, autoLogin bool, err error) {
	// 尝试唤醒登陆
	var cachedInfo CachedInfo
	cachedInfo, err = r.Client.GetCachedInfo(r.WxID)
	if err == nil && cachedInfo.Wxid != "" {
		_, err = r.Client.LoginTwiceAutoAuth(r.WxID)
		if err == nil {
			autoLogin = true
			return
		}
		// 唤醒登陆
		var resp QrCode
		resp, err = r.Client.AwakenLogin(r.WxID, r.DeviceName)
		if err != nil {
			// 如果唤醒失败，尝试获取二维码
			uuid, err = r.GetQrCode()
			return
		}
		if resp.Uuid == "" {
			// 如果唤醒失败，尝试获取二维码
			uuid, err = r.GetQrCode()
			return
		}
		// 唤醒登陆成功
		uuid = resp.Uuid
		awkenLogin = true
		return
	}
	// 二维码登陆
	uuid, err = r.GetQrCode()
	return
}

func (r *Robot) AtListDecoder(xmlStr string) string {
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ""
		}
		switch el := tok.(type) {
		case xml.StartElement:
			if el.Name.Local == "atuserlist" {
				var content string
				decoder.DecodeElement(&content, &el)
				return content
			}
		}
	}
	return ""
}

func (r *Robot) SyncMessage() (SyncMessage, error) {
	return r.Client.SyncMessage(r.WxID)
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

func (r *Robot) Heartbeat() error {
	return r.Client.Heartbeat(r.WxID)
}

func (r *Robot) GetContactList() (wxids []string, err error) {
	return r.Client.GetContactList(r.WxID)
}

func (r *Robot) GetContactDetail(requestWxids []string) (contactList []Contact, err error) {
	return r.Client.GetContactDetail(r.WxID, requestWxids)
}
