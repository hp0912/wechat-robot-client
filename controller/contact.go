package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type Contact struct {
}

func NewContactController() *Contact {
	return &Contact{}
}

func (ct *Contact) SyncContact(c *gin.Context) {
	resp := appx.NewResponse(c)
	err := service.NewContactService(c).SyncContact(false)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (ct *Contact) GetContacts(c *gin.Context) {
	var req dto.ContactListRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	pager := appx.InitPager(c)
	list, total, err := service.NewContactService(c).GetContacts(req, pager)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponseList(list, total)
}

func (ct *Contact) FriendPassVerify(c *gin.Context) {
	var req dto.FriendPassVerifyRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewContactService(c).FriendPassVerify(req.SystemMessageID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (ct *Contact) FriendDelete(c *gin.Context) {
	var req dto.FriendDeleteRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewContactService(c).FriendDelete(req.ContactID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
