package robot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
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
	client      *resty.Client
	Domain      WechatDomain
	limiter     *rate.Limiter
	autoLimiter *rate.Limiter
}

func NewClient(domain WechatDomain) *Client {
	return &Client{
		client:      resty.New(),
		Domain:      domain,
		limiter:     rate.NewLimiter(rate.Every(time.Second), 1),
		autoLimiter: rate.NewLimiter(rate.Every(15*time.Second), 1),
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

func (c *Client) GetProfile(wxid string) (resp GetProfileResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[GetProfileResponse]
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

func (c *Client) GetCachedInfo(wxid string) (resp LoginData, err error) {
	var result ClientResponse[LoginData]
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

func (c *Client) BaseResponseErrCheck(baseResponse *BaseResponse) (err error) {
	if baseResponse != nil && baseResponse.Ret != 0 {
		if baseResponse.ErrMsg != nil && baseResponse.ErrMsg.String != nil && *baseResponse.ErrMsg.String != "" {
			err = fmt.Errorf(*baseResponse.ErrMsg.String)
			return
		}
	}
	return
}

func (c *Client) LoginTwiceAutoAuth(wxid string) (err error) {
	var result ClientResponse[UnifyAuthResponse]
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

func (c *Client) AwakenLogin(wxid string) (resp QrCode, err error) {
	var result ClientResponse[QrCode]
	var httpResp *resty.Response
	httpResp, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(AwakenLoginRequest{
			Wxid: wxid,
		}).
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

func (c *Client) GetQrCode(loginType, deviceId, deviceName string) (resp GetQRCode, err error) {
	var result ClientResponse[GetQRCode]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(LoginGetQRRequest{
			DeviceID:   deviceId,
			DeviceName: deviceName,
			LoginType:  loginType,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginGetQR))
	if err = result.CheckError(err); err != nil {
		return
	}
	result.Data.Data62 = result.Data62
	resp = result.Data
	return
}

func (c *Client) LoginGetQRMac(deviceId, deviceName string) (resp GetQRCode, err error) {
	var result ClientResponse[GetQRCode]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(LoginGetQRRequest{
			DeviceID:   deviceId,
			DeviceName: deviceName,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginGetQRMac))
	if err = result.CheckError(err); err != nil {
		return
	}
	result.Data.Data62 = result.Data62
	resp = result.Data
	return
}

func (c *Client) CheckLoginUuid(uuid string) (resp CheckUuid, err error) {
	var result ClientResponse[json.RawMessage]
	_, err = c.client.R().
		SetResult(&result).
		SetQueryParam("uuid", uuid).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginCheckQR))
	if err = result.CheckError(err); err != nil {
		return
	}
	// 先尝试解析为 CheckUuid 结构体（协议有时候直接返回无规则字符串 ticket）
	if err = json.Unmarshal(result.Data, &resp); err == nil {
		return
	}
	var ticket string
	if err = json.Unmarshal(result.Data, &ticket); err == nil {
		resp = CheckUuid{Uuid: uuid, Ticket: ticket}
		return
	}
	err = fmt.Errorf("无法解析响应数据: %s", string(result.Data))
	return
}

func (c *Client) LoginYPayVerificationcode(req VerificationCodeRequest) (err error) {
	var result ClientResponse[struct{}]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginYPayVerificationcode))
	if err = result.CheckError(err); err != nil {
		return
	}
	return
}

func (c *Client) LoginNewDeviceVerify(ticket string) (resp SilderOCR, err error) {
	var result ClientResponse[SilderOCR]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"ticket": ticket,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginNewDeviceVerify))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) LoginGet62Data(wxid string) (resp string, err error) {
	var result ClientResponse[string]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("wxid", wxid).
		SetResult(&result).
		Get(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginGet62Data))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) LoginGetA16Data(wxid string) (resp string, err error) {
	var result ClientResponse[string]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParam("wxid", wxid).
		SetResult(&result).
		Get(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginGetA16Data))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) LoginData62SMSApply(req Data62LoginRequest) (resp UnifyAuthResponse, err error) {
	var result ClientResponse[UnifyAuthResponse]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginData62SMSApply))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

// LoginData62SMSAgain 重新发送验证码
func (c *Client) LoginData62SMSAgain(req LoginData62SMSAgainRequest) (resp string, err error) {
	var result ClientResponse[string]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginData62SMSAgain))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

// LoginData62SMSVerify 短信验证
func (c *Client) LoginData62SMSVerify(req LoginData62SMSVerifyRequest) (resp string, err error) {
	var result ClientResponse[string]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginData62SMSVerify))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) LoginA16Data1(req A16LoginRequest) (resp UnifyAuthResponse, err error) {
	var result ClientResponse[UnifyAuthResponse]
	_, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginA16Data1))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
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

// AutoHeartBeat 自动心跳，包括自动同步消息
func (c *Client) AutoHeartBeat(wxid string) (err error) {
	var result ClientResponse[struct{}]
	_, err = c.client.R().
		SetResult(&result).
		SetQueryParam("wxid", wxid).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginAutoHeartBeat))
	err = result.CheckError(err)
	return
}

// CloseAutoHeartBeat 关闭自动心跳
func (c *Client) CloseAutoHeartBeat(wxid string) (err error) {
	var result ClientResponse[struct{}]
	_, err = c.client.R().
		SetResult(&result).
		SetQueryParam("wxid", wxid).
		Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), LoginCloseAutoHeartBeat))
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
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSync))
	if err = result.CheckError(err); err != nil {
		return
	}
	messageResponse = result.Data
	return
}

// MessageRevoke 撤回消息
func (c *Client) MessageRevoke(req MessageRevokeRequest) (err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
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
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
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

func (c *Client) MsgSendGroupMassMsgText(req MsgSendGroupMassMsgTextRequest) (newMessages MsgSendGroupMassMsgTextResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[MsgSendGroupMassMsgTextResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSendGroupMassMsgText))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	newMessages = result.Data
	return
}

// MsgUploadImg 发送图片消息
func (c *Client) MsgUploadImg(wxid, toWxid, base64 string) (imageMessage MsgUploadImgResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
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
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
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

// MsgSendVoice 发送音频消息
func (c *Client) MsgSendVoice(req MsgSendVoiceRequest) (voiceMessage MsgSendVoiceResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[MsgSendVoiceResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSendVoice))
	if err = result.CheckError(err); err != nil {
		return
	}
	voiceMessage = result.Data
	return
}

// ToolsSendFile 上传文件
func (c *Client) ToolsSendFile(req SendFileMessageRequest, file io.Reader, fileHeader *multipart.FileHeader) (fileMessage *SendFileMessageResponse, err error) {
	if req.StartPos == 0 {
		if err = c.limiter.Wait(context.Background()); err != nil {
			return
		}
	}

	var requestBody bytes.Buffer
	var part io.Writer
	writer := multipart.NewWriter(&requestBody)

	// 分片文件字段名与前端一致: chunk
	part, err = writer.CreateFormFile("chunk", fileHeader.Filename)
	if err != nil {
		return
	}
	if _, err = io.Copy(part, file); err != nil {
		return
	}
	// 追加其他字段
	if err = writer.WriteField("Wxid", req.Wxid); err != nil {
		return
	}
	if err = writer.WriteField("ClientAppDataId", req.ClientAppDataId); err != nil {
		return
	}
	if err = writer.WriteField("FileMD5", req.FileMD5); err != nil {
		return
	}
	if err = writer.WriteField("TotalLen", strconv.FormatInt(req.TotalLen, 10)); err != nil {
		return
	}
	if err = writer.WriteField("StartPos", strconv.FormatInt(req.StartPos, 10)); err != nil {
		return
	}
	if err = writer.WriteField("TotalChunks", strconv.FormatInt(req.TotalChunks, 10)); err != nil {
		return
	}
	if err = writer.Close(); err != nil {
		return
	}
	var robotRequest *http.Request
	var robotResp *http.Response
	robotRequest, err = http.NewRequest("POST", fmt.Sprintf("%s%s", c.Domain.BasePath(), ToolsUploadAppAttachStream), &requestBody)
	if err != nil {
		return
	}
	robotRequest.Header.Set("Content-Type", writer.FormDataContentType())
	robotClient := &http.Client{}
	robotResp, err = robotClient.Do(robotRequest)
	if err != nil {
		return
	}
	defer robotResp.Body.Close()

	if robotResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("上传文件请求失败，状态码: %d", robotResp.StatusCode)
		return
	}

	var respBody []byte
	respBody, err = io.ReadAll(robotResp.Body)
	if err != nil {
		return
	}

	var result ClientResponse[SendFileMessageResponse]
	if err = json.Unmarshal(respBody, &result); err != nil {
		return
	}

	if err = result.CheckError(nil); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}

	fileMessage = &result.Data

	return
}

// SendApp 发送App消息
func (c *Client) SendApp(req SendAppRequest) (appMessage SendAppResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[SendAppResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSendApp))
	if err = result.CheckError(err); err != nil {
		return
	}
	appMessage = result.Data
	return
}

// SendEmoji 发送表情消息
func (c *Client) SendEmoji(req SendEmojiRequest) (emojiMessage SendEmojiResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[SendEmojiResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSendEmoji))
	if err = result.CheckError(err); err != nil {
		return
	}
	emojiMessage = result.Data
	return
}

// ShareLink 发送分享链接消息
func (c *Client) ShareLink(req ShareLinkRequest) (shareLinkMessage ShareLinkResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[ShareLinkResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgShareLink))
	if err = result.CheckError(err); err != nil {
		return
	}
	shareLinkMessage = result.Data
	return
}

// SendCDNFile 转发文件消息（转发，并非上传）
func (c *Client) SendCDNFile(req SendCDNAttachmentRequest) (cdnFileMessage SendCDNFileResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[SendCDNFileResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSendCDNFile))
	if err = result.CheckError(err); err != nil {
		return
	}
	cdnFileMessage = result.Data
	return
}

// SendCDNImg 转发图片消息（转发，并非上传）
func (c *Client) SendCDNImg(req SendCDNAttachmentRequest) (cdnImageMessage SendCDNImgResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[SendCDNImgResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSendCDNImg))
	if err = result.CheckError(err); err != nil {
		return
	}
	cdnImageMessage = result.Data
	return
}

// SendCDNVideo 转发视频消息（转发，并非上传）
func (c *Client) SendCDNVideo(req SendCDNAttachmentRequest) (cdnVideoMessage SendCDNVideoResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[SendCDNVideoResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), MsgSendCDNVideo))
	if err = result.CheckError(err); err != nil {
		return
	}
	cdnVideoMessage = result.Data
	return
}

func (c *Client) FriendGetFriendstate(Wxid, UserName string) (resp MMBizJsApiGetUserOpenIdResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[MMBizJsApiGetUserOpenIdResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(map[string]string{
			"Wxid":     Wxid,
			"UserName": UserName,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendGetFriendstate))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

func (c *Client) FriendSearch(req FriendSearchRequest) (resp SearchContactResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[SearchContactResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendSearch))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

func (c *Client) FriendSendRequest(req FriendSendRequestParam) (resp VerifyUserResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[VerifyUserResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendSendRequest))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

func (c *Client) FriendSetRemarks(wxid, toWxid, remarks string) (resp OplogResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[OplogResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(FriendSetRemarksRequest{
			Wxid:    wxid,
			ToWxid:  toWxid,
			Remarks: remarks,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendSetRemarks))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
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
	return
}

func (c *Client) GetContactDetail(wxid, chatRoomID string, towxids []string) (resp GetContactResponse, err error) {
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
			ChatRoom: chatRoomID,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendGetContactDetail))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

// FriendPassVerify 通过好友验证
func (c *Client) FriendPassVerify(req FriendPassVerifyRequest) (verifyUserResponse VerifyUserResponse, err error) {
	if err = c.autoLimiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[VerifyUserResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendPassVerify))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	verifyUserResponse = result.Data
	return
}

func (c *Client) FriendDelete(Wxid, ToWxid string) (oplogResponse OplogResponse, err error) {
	var result ClientResponse[OplogResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(FriendDeleteRequest{
			Wxid:   Wxid,
			ToWxid: ToWxid,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendDelete))
	if err = result.CheckError(err); err != nil {
		return
	}
	oplogResponse = result.Data
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

func (c *Client) CreateChatRoom(wxid string, contactIDs []string) (createChatRoomResp CreateChatRoomResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[CreateChatRoomResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(CreateChatRoomRequest{
			Wxid:    wxid,
			ToWxids: strings.Join(contactIDs, ","),
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupCreateChatRoom))
	if err = result.CheckError(err); err != nil {
		return
	}
	createChatRoomResp = result.Data
	return
}

func (c *Client) GroupAddChatRoomMember(wxid, chatRoomName string, contactIDs []string) (memberResp InviteChatRoomMemberResponse, err error) {
	if err = c.autoLimiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[InviteChatRoomMemberResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(InviteChatRoomMemberRequest{
			Wxid:         wxid,
			ChatRoomName: chatRoomName,
			ToWxids:      strings.Join(contactIDs, ","),
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupAddChatRoomMember))
	if err = result.CheckError(err); err != nil {
		return
	}
	memberResp = result.Data
	return
}

func (c *Client) GroupInviteChatRoomMember(wxid, chatRoomName string, contactIDs []string) (memberResp InviteChatRoomMemberResponse, err error) {
	if err = c.autoLimiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[InviteChatRoomMemberResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(InviteChatRoomMemberRequest{
			Wxid:         wxid,
			ChatRoomName: chatRoomName,
			ToWxids:      strings.Join(contactIDs, ","),
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupInviteChatRoomMember))
	if err = result.CheckError(err); err != nil {
		return
	}
	memberResp = result.Data
	return
}

func (c *Client) GroupConsentToJoin(wxid, Url string) (QID string, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[string]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(ConsentToJoinRequest{
			Wxid: wxid,
			Url:  Url,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupConsentToJoin))
	if err = result.CheckError(err); err != nil {
		return
	}
	QID = result.Data
	return
}

func (c *Client) GetChatRoomMemberDetail(wxid, QID string) (chatRoomMember []ChatRoomMember, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[ChatRoomMemberDetail]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(ChatRoomRequestBase{
			Wxid: wxid,
			QID:  QID,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupGetChatRoomMemberDetail))
	if err = result.CheckError(err); err != nil {
		return
	}
	chatRoomMember = result.Data.NewChatroomData.ChatRoomMember
	return
}

func (c *Client) GroupSetChatRoomName(wxid, QID, Content string) (err error) {
	var result ClientResponse[OplogResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(OperateChatRoomInfoParam{
			Wxid:    wxid,
			QID:     QID,
			Content: Content,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupSetChatRoomName))
	if err = result.CheckError(err); err != nil {
		return
	}
	return
}

func (c *Client) GroupSetChatRoomRemarks(wxid, QID, Content string) (err error) {
	var result ClientResponse[OplogResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(OperateChatRoomInfoParam{
			Wxid:    wxid,
			QID:     QID,
			Content: Content,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupSetChatRoomRemarks))
	if err = result.CheckError(err); err != nil {
		return
	}
	return
}

func (c *Client) GroupSetChatRoomAnnouncement(wxid, QID, Content string) (err error) {
	var result ClientResponse[OplogResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(OperateChatRoomInfoParam{
			Wxid:    wxid,
			QID:     QID,
			Content: Content,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupSetChatRoomAnnouncement))
	if err = result.CheckError(err); err != nil {
		return
	}
	return
}

func (c *Client) GroupDelChatRoomMember(wxid, QID string, ToWxids []string) (err error) {
	var result ClientResponse[DelChatRoomMemberResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(DelChatRoomMemberRequest{
			Wxid:         wxid,
			ChatRoomName: QID,
			ToWxids:      strings.Join(ToWxids, ","),
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupDelChatRoomMember))
	if err = result.CheckError(err); err != nil {
		return
	}
	return
}

func (c *Client) GroupQuit(wxid, QID string) (err error) {
	var result ClientResponse[OplogResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(ChatRoomRequestBase{
			Wxid: wxid,
			QID:  QID,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), GroupQuit))
	if err = result.CheckError(err); err != nil {
		return
	}
	return
}

// 朋友圈接口

// FriendCircleComment 朋友圈评论
func (c *Client) FriendCircleComment(req FriendCircleCommentRequest) (resp SnsCommentResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[SnsCommentResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleComment))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

// FriendCircleGetDetail 获取特定人朋友圈
func (c *Client) FriendCircleGetDetail(req FriendCircleGetDetailRequest) (resp SnsUserPageResponse, err error) {
	var result ClientResponse[SnsUserPageResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleGetDetail))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

// FriendCircleGetIdDetail 获取特定ID详情内容
func (c *Client) FriendCircleGetIdDetail(req FriendCircleGetIdDetailRequest) (resp SnsObjectDetailResponse, err error) {
	var result ClientResponse[SnsObjectDetailResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleGetIdDetail))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

// FriendCircleGetList 获取朋友圈列表
func (c *Client) FriendCircleGetList(wxid, Fristpagemd5 string, Maxid string) (Moments GetListResponse, err error) {
	var result ClientResponse[GetListResponse]
	var maxID uint64

	maxID, err = strconv.ParseUint(Maxid, 10, 64)
	if err != nil {
		return
	}

	_, err = c.client.R().
		SetResult(&result).
		SetBody(GetListRequest{
			Wxid:         wxid,
			Fristpagemd5: Fristpagemd5,
			Maxid:        maxID,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleGetList))
	if err = result.CheckError(err); err != nil {
		return
	}
	Moments = result.Data
	return
}

// FriendCircleDownFriendCircleMedia 下载朋友圈视频
func (c *Client) FriendCircleDownFriendCircleMedia(wxid, Url, Key string) (mediaBase64 string, err error) {
	var result ClientResponse[string]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(DownFriendCircleMediaRequest{
			Wxid: wxid,
			Url:  Url,
			Key:  Key,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleDownFriendCircleMedia))
	if err = result.CheckError(err); err != nil {
		return
	}
	mediaBase64 = result.Data
	return
}

// 朋友圈图片上传
func (c *Client) FriendCircleUpload(wxid string, base64 string) (resp FriendCircleUploadResponse, err error) {
	var result ClientResponse[FriendCircleUploadResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(FriendCircleUploadRequest{
			Wxid:   wxid,
			Base64: base64,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleUpload))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

// 朋友圈视频上传
func (c *Client) FriendCircleCdnSnsUploadVideo(req FriendCircleCdnSnsUploadVideoRequest) (resp CdnSnsVideoUploadResponse, err error) {
	var result ClientResponse[CdnSnsVideoUploadResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleCdnSnsUploadVideo))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

// 朋友圈操作
func (c *Client) FriendCircleOperation(req FriendCircleOperationRequest) (resp SnsObjectOpResponse, err error) {
	if err = c.limiter.Wait(context.Background()); err != nil {
		return
	}
	var result ClientResponse[SnsObjectOpResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleOperation))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

// 朋友圈权限设置
func (c *Client) FriendCirclePrivacySettings(req FriendCirclePrivacySettingsRequest) (resp OplogResponse, err error) {
	var result ClientResponse[OplogResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCirclePrivacySettings))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

// 发布朋友圈
func (c *Client) FriendCircleMessages(req FriendCircleMessagesRequest) (resp FriendCircleMessagesResponse, err error) {
	var result ClientResponse[FriendCircleMessagesResponse]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(req).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleMessages))
	if err = result.CheckError(err); err != nil {
		err2 := c.BaseResponseErrCheck(result.Data.BaseResponse)
		if err2 != nil {
			err = err2
			return
		}
		return
	}
	resp = result.Data
	return
}

// FriendCircleMmSnsSync 同步朋友圈
func (c *Client) FriendCircleMmSnsSync(wxid, synckey string) (resp SyncMessage, err error) {
	var result ClientResponse[SyncMessage]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(map[string]string{
			"Wxid":    wxid,
			"Synckey": synckey,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), FriendCircleMmSnsSync))
	if err = result.CheckError(err); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) WxappQrcodeAuthLogin(wxid, Url string) (err error) {
	var result ClientResponse[struct{}]
	_, err = c.client.R().
		SetResult(&result).
		SetBody(map[string]string{
			"Wxid": wxid,
			"Url":  Url,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BasePath(), WxappQrcodeAuthLogin))
	if err = result.CheckError(err); err != nil {
		return
	}
	return
}
