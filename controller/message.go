package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type Message struct{}

func NewMessageController() *Message {
	return &Message{}
}

func (m *Message) MessageRevoke(c *gin.Context) {
	var req dto.MessageCommonRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMessageService(c).MessageRevoke(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}

func (m *Message) SendTextMessage(c *gin.Context) {
	var req dto.SendTextMessageRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewMessageService(c).SendTextMessage(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
