package controller

import (
	"errors"
	"net/http"
	"wechat-robot-client/dto"
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

func (d *Robot) SyncContact(c *gin.Context) {
	resp := appx.NewResponse(c)
	err := service.NewRobotService(c).SyncContact(false)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (d *Robot) GetContacts(c *gin.Context) {
	var req dto.ContactListRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	pager := appx.InitPager(c)
	list, total, err := service.NewRobotService(c).GetContacts(req, pager)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponseList(list, total)
}

func (d *Robot) SyncChatRoomMember(c *gin.Context) {
	var req dto.SyncChatRoomMemberRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	service.NewRobotService(c).SyncChatRoomMember(req.ChatRoomID)
	resp.ToResponse(nil)
}

func (d *Robot) GetChatRoomMembers(c *gin.Context) {
	var req dto.ChatRoomMemberRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	pager := appx.InitPager(c)
	list, total, err := service.NewRobotService(c).GetChatRoomMembers(req, pager)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponseList(list, total)
}

func (d *Robot) GetChatHistory(c *gin.Context) {
	var req dto.ChatHistoryRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	pager := appx.InitPager(c)
	list, total, err := service.NewRobotService(c).GetChatHistory(req, pager)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponseList(list, total)
}

func (d *Robot) Login(c *gin.Context) {
	resp := appx.NewResponse(c)
	uuid, awkenLogin, autoLogin, err := service.NewRobotService(c).Login()
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
