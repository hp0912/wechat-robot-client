package controller

import (
	"log"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type WechatServerCallback struct {
}

func NewWechatServerCallbackController() *WechatServerCallback {
	return &WechatServerCallback{}
}

func (ct *WechatServerCallback) SyncMessageCallback(c *gin.Context) {
	wechatID := c.Param("wechatID")
	log.Printf("Received SyncMessageCallback for wechatID: %s", wechatID)
	var req robot.ClientResponse[robot.SyncMessage]
	resp := appx.NewResponse(c)
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(err)
		return
	}
	service.NewLoginService(c).SyncMessageCallback(wechatID, req.Data)

	resp.ToResponse(nil)
}

func (ct *WechatServerCallback) LogoutCallback(c *gin.Context) {
	wechatID := c.Param("wechatID")
	log.Printf("Received LogoutCallback for wechatID: %s", wechatID)
	err := service.NewLoginService(c).LogoutCallback(wechatID)
	resp := appx.NewResponse(c)
	if err != nil {
		log.Printf("LogoutCallback failed: %v\n", err)
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
