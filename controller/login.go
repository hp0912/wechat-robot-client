package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/pkg/robot"
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
	var req struct {
		LoginType string `json:"login_type"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	data, err := service.NewLoginService(c).Login(req.LoginType)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
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
	err := service.NewLoginService(c).LoginYPayVerificationcode(robot.VerificationCodeRequest{
		Uuid:   req.Uuid,
		Code:   req.Code,
		Ticket: req.Ticket,
		Data62: req.Data62,
	})
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (lg *Login) LoginNewDeviceVerify(c *gin.Context) {
	var req dto.NewDeviceVerifyRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	data, err := service.NewLoginService(c).LoginNewDeviceVerify(req.Ticket)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
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
