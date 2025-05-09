package controller

import (
	"errors"
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
