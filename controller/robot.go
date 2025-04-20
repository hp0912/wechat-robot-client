package controller

import (
	"net/http"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"

	"github.com/gin-gonic/gin"
)

type Robot struct {
}

func NewRobotController() *Robot {
	return &Robot{}
}

func (d *Robot) Probe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (d *Robot) IsRunning(c *gin.Context) {
	resp := appx.NewResponse(c)
	resp.ToResponse(vars.RobotRuntime.IsRunning())
}

func (d *Robot) IsLoggedIn(c *gin.Context) {
	resp := appx.NewResponse(c)
	resp.ToResponse(vars.RobotRuntime.IsLoggedIn())
}

func (d *Robot) Login(c *gin.Context) {
	resp := appx.NewResponse(c)
	uuid, qrcode, err := service.NewRobotService(c).Login()
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(gin.H{
		"qrcode": qrcode,
		"uuid":   uuid,
	})
}

func (d *Robot) LoginCheck(c *gin.Context) {
	resp := appx.NewResponse(c)
	loggedIn, err := service.NewRobotService(c).LoginCheck()
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(gin.H{
		"logged_in": loggedIn,
	})
}

func (d *Robot) Logout(c *gin.Context) {
	resp := appx.NewResponse(c)
	err := service.NewRobotService(c).Logout()
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
