package controller

import (
	"errors"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type OSSSettings struct{}

func NewOSSSettingsController() *OSSSettings {
	return &OSSSettings{}
}

func (s *OSSSettings) GetOSSSettings(c *gin.Context) {
	resp := appx.NewResponse(c)
	data, err := service.NewOSSSettingService(c).GetOSSSettingService()
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(data)
}

func (s *OSSSettings) SaveOSSSettings(c *gin.Context) {
	var req model.OSSSettings
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewOSSSettingService(c).SaveOSSSettingService(&req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
