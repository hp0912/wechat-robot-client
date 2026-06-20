package controller

import (
	"errors"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type OfficialAccount struct{}

func NewOfficialAccountController() *OfficialAccount {
	return &OfficialAccount{}
}

func (s *OfficialAccount) GetAppMsgExt(c *gin.Context) {
	var req struct {
		URL string `form:"url" json:"url"`
	}
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	msgext, err := service.NewOfficialAccountService(c).GetAppMsgExt(req.URL)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(msgext)
}
