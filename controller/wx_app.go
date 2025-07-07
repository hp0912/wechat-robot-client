package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type WXApp struct{}

func NewWXAppController() *WXApp {
	return &WXApp{}
}

func (m *WXApp) WxappQrcodeAuthLogin(c *gin.Context) {
	var req dto.WxappQrcodeAuthLoginRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewWXAppService(c).WxappQrcodeAuthLogin(req.URL)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
