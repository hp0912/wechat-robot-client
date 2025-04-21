package controller

import (
	"errors"
	"net/http"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

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
	resp.ToResponse(service.NewRobotService(c).IsRunning())
}

func (d *Robot) IsLoggedIn(c *gin.Context) {
	resp := appx.NewResponse(c)
	resp.ToResponse(service.NewRobotService(c).IsLoggedIn())
}

func (d *Robot) Login(c *gin.Context) {
	resp := appx.NewResponse(c)
	uuid, awken, err := service.NewRobotService(c).Login()
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(gin.H{
		"uuid":  uuid,
		"awken": awken,
	})
}

func (d *Robot) LoginCheck(c *gin.Context) {
	var req struct {
		Uuid string `json:"uuid" binding:"required"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	data, err := service.NewRobotService(c).LoginCheck(req.Uuid)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
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
