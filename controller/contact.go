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

func (ct *Contact) FriendSearch(c *gin.Context) {
	var req dto.FriendSearchRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	friend, err := service.NewContactService(c).FriendSearch(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(friend)
}

func (ct *Contact) FriendSendRequest(c *gin.Context) {
	var req dto.FriendSendRequestRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewContactService(c).FriendSendRequest(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (ct *Contact) FriendSendRequestFromChatRoom(c *gin.Context) {
	var req dto.FriendSendRequestFromChatRoomRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewContactService(c).FriendSendRequestFromChatRoom(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (ct *Contact) FriendSetRemarks(c *gin.Context) {
	var req dto.FriendSetRemarksRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewContactService(c).FriendSetRemarks(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
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
