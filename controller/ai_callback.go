package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type AICallback struct {
}

func NewAICallbackController() *AICallback {
	return &AICallback{}
}

func (a *AICallback) DoubaoTTS(c *gin.Context) {
	var req dto.DoubaoTTSCallbackRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewAICallbackService(c).DoubaoTTS(req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
