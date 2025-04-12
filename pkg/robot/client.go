package robot

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type ClientResponse[T any] struct {
	Success bool   `json:"Success"`
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	Data    T      `json:"Data"`
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
	resp, err := c.client.R().Get(fmt.Sprintf("%s%s", c.Domain.BaseHost(), IsRunning))
	if err != nil || resp.StatusCode() != http.StatusOK {
		log.Printf("Error checking if robot is running: %v, http code: %d", err, resp.StatusCode())
		return false
	}
	return resp.String() == "OK"
}

type UserProfileRequest struct {
	Wxid string `json:"wxid"`
}

type UserProfile struct {
	BaseResponse BaseResponse `json:"baseResponse"`
	UserInfo     UserInfo     `json:"userInfo"`
	UserInfoExt  UserInfoExt  `json:"userInfoExt"`
}

func (c *Client) GetProfile(wxid string) (*ClientResponse[UserProfile], error) {
	var userProfile ClientResponse[UserProfile]
	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(UserProfileRequest{
			Wxid: wxid,
		}).
		SetResult(&userProfile).
		Post(fmt.Sprintf("%s%s", c.Domain.BaseHost(), GetProfile))
	if err != nil {
		log.Printf("获取登陆用户信息异常: %v", err)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Printf("获取登陆用户信息状态码不为200: %d", resp.StatusCode())
		return nil, errors.New("获取登陆用户信息状态码不为200")
	}
	return &userProfile, nil
}
