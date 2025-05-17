package controller

import (
	"errors"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type GlobalSettings struct {
}

func NewGlobalSettingsController() *GlobalSettings {
	return &GlobalSettings{}
}

func (ct *GlobalSettings) GetGlobalSettings(c *gin.Context) {
	resp := appx.NewResponse(c)
	globalSettings := service.NewGlobalSettingsService(c).GetGlobalSettings()
	if globalSettings == nil {
		resp.ToErrorResponse(errors.New("获取全局设置失败"))
		return
	}
	resp.ToResponse(globalSettings)
}

func (ct *GlobalSettings) SaveGlobalSettings(c *gin.Context) {
	var req model.GlobalSettings
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	service.NewGlobalSettingsService(c).SaveGlobalSettings(&req)
	resp.ToResponse(nil)
}
