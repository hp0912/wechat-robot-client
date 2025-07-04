package controller

import (
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type SystemMessage struct{}

func NewSystemMessageController() *SystemMessage {
	return &SystemMessage{}
}

func (m *SystemMessage) GetRecentMonthMessages(c *gin.Context) {
	resp := appx.NewResponse(c)
	data, err := service.NewSystemMessageService(c).GetRecentMonthMessages()
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}
