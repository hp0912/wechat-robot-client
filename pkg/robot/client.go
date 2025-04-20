package robot

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// 错误码对应的含义，0表示正常
var codeMap = map[int]string{
	-1:  "参数错误",
	-2:  "其他错误",
	-3:  "序列化错误",
	-4:  "反序列化错误",
	-5:  "MMTLS初始化错误",
	-6:  "收到的数据包长度错误",
	-7:  "已退出登录",
	-8:  "链接过期",
	-9:  "解析数据包错误",
	-10: "数据库错误",
	-11: "登陆异常",
	-12: "操作过于频繁",
	-13: "上传失败",
}

type ClientResponse[T any] struct {
	Success bool   `json:"Success"`
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	Data    T      `json:"Data"`
}

func (c ClientResponse[T]) IsSuccess() bool {
	return c.Code == 0
}

func (c ClientResponse[T]) CheckError(err error, resp *resty.Response) error {
	if err != nil {
		return err
	}
	if errMsg, ok := codeMap[c.Code]; ok {
		return fmt.Errorf("[%d] %s - %s", c.Code, errMsg, c.Message)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("http error: %d - %s", resp.StatusCode(), resp.String())
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
	resp, err := c.client.R().Get(fmt.Sprintf("%s%s", c.Domain.BaseHost(), IsRunningPath))
	if err != nil || resp.StatusCode() != http.StatusOK {
		log.Printf("Error checking if robot is running: %v, http code: %d", err, resp.StatusCode())
		return false
	}
	return resp.String() == "OK"
}

type CommonRequest struct {
	Wxid string `json:"Wxid"`
}

func (c *Client) GetProfile(wxid string) (resp UserProfile, err error) {
	var result ClientResponse[UserProfile]
	var httpResp *resty.Response
	httpResp, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(CommonRequest{
			Wxid: wxid,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BaseHost(), GetProfilePath))
	if err = result.CheckError(err, httpResp); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) GetCachedInfo(wxid string) (resp CachedInfo, err error) {
	var result ClientResponse[CachedInfo]
	var httpResp *resty.Response
	httpResp, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(CommonRequest{
			Wxid: wxid,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BaseHost(), GetCachedInfoPath))
	if err = result.CheckError(err, httpResp); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) AwakenLogin(wxid string) (resp AwakenLogin, err error) {
	var result ClientResponse[AwakenLogin]
	var httpResp *resty.Response
	httpResp, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(CommonRequest{
			Wxid: wxid,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BaseHost(), AwakenLoginPath))
	if err = result.CheckError(err, httpResp); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) GetQrCode(deviceId, deviceName string) (resp GetQRCode, err error) {
	var result ClientResponse[GetQRCode]
	var httpResp *resty.Response
	httpResp, err = c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"DeviceID":   deviceId,
			"DeviceName": deviceName,
		}).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.Domain.BaseHost(), GetQrCodePath))
	if err = result.CheckError(err, httpResp); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) CheckLoginUuid(uuid string) (resp CheckUuid, err error) {
	var result ClientResponse[CheckUuid]
	var httpResp *resty.Response
	httpResp, err = c.client.R().
		SetResult(&result).
		SetBody(map[string]string{
			"Uuid": uuid,
		}).Post(fmt.Sprintf("%s%s", c.Domain.BaseHost(), CheckUuidPath))
	if err = result.CheckError(err, httpResp); err != nil {
		return
	}
	resp = result.Data
	return
}

func (c *Client) Logout(wxid string) (err error) {
	var result ClientResponse[struct{}]
	var httpResp *resty.Response
	httpResp, err = c.client.R().SetResult(&result).SetBody(CommonRequest{
		Wxid: wxid,
	}).Post(fmt.Sprintf("%s%s", c.Domain.BaseHost(), LogoutPath))
	err = result.CheckError(err, httpResp)
	return
}
