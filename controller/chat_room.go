package controller

import (
	"errors"
	"unicode/utf8"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type ChatRoom struct {
}

func NewChatRoomController() *ChatRoom {
	return &ChatRoom{}
}

func (cr *ChatRoom) SyncChatRoomMember(c *gin.Context) {
	var req dto.SyncChatRoomMemberRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	service.NewChatRoomService(c).SyncChatRoomMember(req.ChatRoomID)
	resp.ToResponse(nil)
}

func (cr *ChatRoom) GetChatRoomMembers(c *gin.Context) {
	var req dto.ChatRoomMemberRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	pager := appx.InitPager(c)
	list, total, err := service.NewChatRoomService(c).GetChatRoomMembers(req, pager)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponseList(list, total)
}

func (cr *ChatRoom) GetNotLeftMembers(c *gin.Context) {
	var req dto.ChatRoomMemberRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	list, err := service.NewChatRoomService(c).GetNotLeftMembers(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(list)
}

func (cr *ChatRoom) CreateChatRoom(c *gin.Context) {
	var req dto.CreateChatRoomRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewChatRoomService(c).CreateChatRoom(req.ContactIDs)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (cr *ChatRoom) InviteChatRoomMember(c *gin.Context) {
	var req dto.InviteChatRoomMemberRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewChatRoomService(c).InviteChatRoomMember(req.ChatRoomID, req.ContactIDs)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (cr *ChatRoom) GroupConsentToJoin(c *gin.Context) {
	var req dto.GroupConsentToJoinRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewChatRoomService(c).GroupConsentToJoin(req.SystemMessageID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (cr *ChatRoom) GroupSetChatRoomName(c *gin.Context) {
	var req dto.ChatRoomOperateRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}

	if utf8.RuneCountInString(req.Content) > 30 {
		resp.ToErrorResponse(errors.New("群名称不能超过30个字符"))
		return
	}
	err := service.NewChatRoomService(c).GroupSetChatRoomName(req.ChatRoomID, req.Content)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (cr *ChatRoom) GroupSetChatRoomRemarks(c *gin.Context) {
	var req dto.ChatRoomOperateRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if utf8.RuneCountInString(req.Content) > 30 {
		resp.ToErrorResponse(errors.New("群备注不能超过30个字符"))
		return
	}
	err := service.NewChatRoomService(c).GroupSetChatRoomRemarks(req.ChatRoomID, req.Content)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (cr *ChatRoom) GroupSetChatRoomAnnouncement(c *gin.Context) {
	var req dto.ChatRoomOperateRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewChatRoomService(c).GroupSetChatRoomAnnouncement(req.ChatRoomID, req.Content)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (cr *ChatRoom) GroupDelChatRoomMember(c *gin.Context) {
	var req dto.DelChatRoomMemberRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewChatRoomService(c).GroupDelChatRoomMember(req.ChatRoomID, req.MemberIDs)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (cr *ChatRoom) GroupQuit(c *gin.Context) {
	var req dto.ChatRoomRequestBase
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewChatRoomService(c).GroupQuit(req.ChatRoomID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
