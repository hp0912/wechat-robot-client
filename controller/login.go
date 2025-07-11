package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type Login struct {
}

func NewLoginController() *Login {
	return &Login{}
}

func (lg *Login) IsRunning(c *gin.Context) {
	resp := appx.NewResponse(c)
	resp.ToResponse(service.NewLoginService(c).IsRunning())
}

func (lg *Login) IsLoggedIn(c *gin.Context) {
	resp := appx.NewResponse(c)
	resp.ToResponse(service.NewLoginService(c).IsLoggedIn())
}

func (lg *Login) Login(c *gin.Context) {
	resp := appx.NewResponse(c)
	uuid, awkenLogin, autoLogin, err := service.NewLoginService(c).Login()
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(gin.H{
		"uuid":        uuid,
		"awken_login": awkenLogin,
		"auto_login":  autoLogin,
	})
}

func (lg *Login) LoginCheck(c *gin.Context) {
	var req struct {
		Uuid string `json:"uuid" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	data, err := service.NewLoginService(c).LoginCheck(req.Uuid)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}

func (lg *Login) LoginYPayVerificationcode(c *gin.Context) {
	var req dto.TFARequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewLoginService(c).LoginYPayVerificationcode(req.Uuid, req.Code, req.Ticket)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (lg *Login) Logout(c *gin.Context) {
	resp := appx.NewResponse(c)
	err := service.NewLoginService(c).Logout()
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
