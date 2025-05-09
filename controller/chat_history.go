package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type ChatHistory struct {
}

func NewChatHistoryController() *ChatHistory {
	return &ChatHistory{}
}

func (ch *ChatHistory) GetChatHistory(c *gin.Context) {
	var req dto.ChatHistoryRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	pager := appx.InitPager(c)
	list, total, err := service.NewChatHistoryService(c).GetChatHistory(req, pager)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponseList(list, total)
}
