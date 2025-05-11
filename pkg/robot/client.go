package robot

import (
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type ClientResponse[T any] struct {
	Success bool   `json:"Success"`
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	Data    T      `json:"Data"`
	Data62  string `json:"Data62"`
	Debug   string `json:"Debug"`
}

func (c ClientResponse[T]) IsSuccess() bool {
	return c.Code == 0
}

func (c ClientResponse[T]) CheckError(err error) error {
	if err != nil {
		return err
	}
	if c.Success {
		return nil
	}
	switch c.Code {
	case 0:
		return nil
	case -7:
		return errors.New("已退出登录")
	case -1, -2, -3, -4, -5, -6, -8, -9, -10, -11, -12, -13:
		return errors.New(c.Message)
	}
	return nil
}

type Client struct {
	client *resty.Client
	Domain WechatDomain
}

func NewClient(domain WechatDomain) *Client {
	return &Client{
		client: resty.New(),
		Domain: domain,
	}
}

func (c *Client) IsRunning() bool {
	timeout := time.Second * 1
	conn, err := net.DialTimeout("tcp", string(c.Domain), timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func (c *Client) GetProfile(wxid string) (resp UserProfile, err error) {
	var result ClientResponse[UserProfile]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("wxid", wxid).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), UserGetContactProfile))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) GetCachedInfo(wxid string) (resp CachedInfo, err error) {
	var result ClientResponse[CachedInfo]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("wxid", wxid).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginGetCacheInfo))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) LoginTwiceAutoAuth(wxid string) (err error) {
	var result ClientResponse[struct{}]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("wxid", wxid).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginTwiceAutoAuth))
	if err = result.CheckError(err); err != nil {
		return
	}
	return
}

func (c *Client) AwakenLogin(wxid string, deviceName string) (resp QrCode, err error) {
	var result ClientResponse[QrCode]
	var httpResp *resty.Response
	httpResp, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("wxid", wxid).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginAwaken))
	if err = result.CheckError(err); err != nil {
		return
	}
	if httpResp.StatusCode() != 200 {
		err = fmt.Errorf("请求失败，状态码：%d", httpResp.StatusCode())
		return
	}
	resp = result.Data
	return
}

func (c *Client) GetQrCode(deviceId, deviceName string) (resp GetQRCode, err error) {
	var result ClientResponse[GetQRCode]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"DeviceID":   deviceId,
			"DeviceName": deviceName,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginGetQR))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) CheckLoginUuid(uuid string) (resp CheckUuid, err error) {
	var result ClientResponse[CheckUuid]
	_, err = c.client.R().
		SetResult(&result).
		SetQueryParam("uuid", uuid).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginCheckQR))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) Logout(wxid string) (err error) {
	var result ClientResponse[struct{}]
	_, err = c.client.R().
		SetResult(&result).
		SetQueryParam("wxid", wxid).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginLogout))
	err = result.CheckError(err)
	return
}

// Heartbeat 手动发起心跳
func (c *Client) Heartbeat(wxid string) (err error) {
	var result ClientResponse[struct{}]
	_, err = c.client.R().
		SetResult(&result).
		SetQueryParam("wxid", wxid).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginHeartBeat))
	err = result.CheckError(err)
	return
}

// SyncMessage 同步消息
func (c *Client) SyncMessage(wxid string) (messageResponse SyncMessage, err error) {
	var result ClientResponse[SyncMessage]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(SyncMessageRequest{
			Wxid:    wxid,
			Scene:   0,
			Synckey: "",
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSyncPath))
	if err = result.CheckError(err); err != nil {
		return
	}
	messageResponse = result.Data
	return
}

// MessageRevoke 撤回消息
func (c *Client) MessageRevoke(req MessageRevokeRequest) (err error) {
	var result ClientResponse[MessageRevokeResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgRevoke))
	if err = result.CheckError(err); err != nil {
		return
	}
	return
}

// SendTextMessage 发送文本消息
func (c *Client) SendTextMessage(req SendTextMessageRequest) (newMessages SendTextMessageResponse, err error) {
	var result ClientResponse[SendTextMessageResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSendTxt))
	if err = result.CheckError(err); err != nil {
		return
	}
	newMessages = result.Data
	return
}

// MsgUploadImg 发送图片消息
func (c *Client) MsgUploadImg(wxid, toWxid, base64 string) (imageMessage MsgUploadImgResponse, err error) {
	var result ClientResponse[MsgUploadImgResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(MsgUploadImgRequest{
			Wxid:   wxid,
			ToWxid: toWxid,
			Base64: base64,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgUploadImg))
	if err = result.CheckError(err); err != nil {
		return
	}
	imageMessage = result.Data
	return
}

// MsgSendVideo 发送视频消息
func (c *Client) MsgSendVideo(req MsgSendVideoRequest) (videoMessage MsgSendVideoResponse, err error) {
	var result ClientResponse[MsgSendVideoResponse]
	req.Base64 = "data:video/mp4;base64," + req.Base64
	req.ImageBase64 = "data:image/jpeg;base64," + req.ImageBase64
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSendVideo))
	if err = result.CheckError(err); err != nil {
		return
	}
	videoMessage = result.Data
	return
}

func (c *Client) GetContactList(wxid string) (wxids []string, err error) {
	var result ClientResponse[GetContactListResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(GetContactListRequest{
			Wxid:                      wxid,
			CurrentChatRoomContactSeq: 0,
			CurrentWxcontactSeq:       0,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendGetContactList))
	if err = result.CheckError(err); err != nil {
		return
	}
	wxids = result.Data.ContactUsernameList
	// 过滤掉系统微信Id
	var specialId = []string{"filehelper", "newsapp", "fmessage", "weibo", "qqmail", "tmessage", "qmessage", "qqsync",
		"floatbottle", "lbsapp", "shakeapp", "medianote", "qqfriend", "readerapp", "blogapp", "facebookapp", "masssendapp",
		"meishiapp", "feedsapp", "voip", "blogappweixin", "weixin", "brandsessionholder", "weixinreminder", "officialaccounts",
		"notification_messages", "wxitil", "userexperience_alarm", "notification_messages", "exmail_tool", "mphelper"}
	wxids = slices.DeleteFunc(wxids, func(id string) bool {
		return slices.Contains(specialId, id) || strings.HasPrefix(id, "gh_") || strings.TrimSpace(id) == ""
	})
	return
}

func (c *Client) GetContactDetail(wxid string, towxids []string) (contactList []Contact, err error) {
	if len(towxids) > 20 {
		err = errors.New("一次最多查询20个联系人")
		return
	}
	var result ClientResponse[GetContactResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(GetContactDetailRequest{
			Wxid:     wxid,
			Towxids:  strings.Join(towxids, ","),
			ChatRoom: "",
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendGetContactDetail))
	if err = result.CheckError(err); err != nil {
		return
	}
	contactList = result.Data.ContactList
	return
}

func (c *Client) GetChatRoomMemberDetail(wxid, QID string) (chatRoomMember []ChatRoomMember, err error) {
	var result ClientResponse[ChatRoomMemberDetail]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(GetChatRoomMemberDetailRequest{
			Wxid: wxid,
			QID:  QID,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupGetChatRoomMemberDetail))
	if err = result.CheckError(err); err != nil {
		return
	}
	chatRoomMember = result.Data.NewChatroomData.ChatRoomMember
	return
}

func (c *Client) CdnDownloadImg(wxid, aeskey, cdnmidimgurl string) (imgbase64 string, err error) {
	var result ClientResponse[DownloadImageDetail]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(CdnDownloadImgRequest{
			Wxid:       wxid,
			FileAesKey: aeskey,
			FileNo:     cdnmidimgurl,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), ToolsCdnDownloadImage))
	if err = result.CheckError(err); err != nil {
		return
	}
	imgbase64 = result.Data.Image
	return
}

func (c *Client) DownloadVideo(req DownloadVideoRequest) (videobase64 string, err error) {
	var result ClientResponse[DownloadVideoDetail]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), ToolsDownloadVideo))
	if err = result.CheckError(err); err != nil {
		return
	}
	videobase64 = result.Data.Data.Buffer
	return
}

func (c *Client) DownloadVoice(req DownloadVoiceRequest) (voicebase64 string, err error) {
	var result ClientResponse[DownloadVoiceDetail]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), ToolsDownloadVoice))
	if err = result.CheckError(err); err != nil {
		return
	}
	voicebase64 = result.Data.Data.Buffer
	return
}

func (c *Client) DownloadFile(req DownloadFileRequest) (filebase64 string, err error) {
	var result ClientResponse[DownloadFileDetail]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), ToolsDownloadFile))
	if err = result.CheckError(err); err != nil {
		return
	}
	filebase64 = result.Data.Data.Buffer
	return
}

// TODO 通过好友请求
