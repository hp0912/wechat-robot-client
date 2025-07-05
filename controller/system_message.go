package controller

import (
	"errors"
	"wechat-robot-client/dto"
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

func (m *SystemMessage) MarkAsReadBatch(c *gin.Context) {
	var req dto.MarkAsReadBatchRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if err := service.NewSystemMessageService(c).MarkAsReadBatch(req.IDs); err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
