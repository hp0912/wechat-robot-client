package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type FriendSettings struct {
}

func NewFriendSettingsController() *FriendSettings {
	return &FriendSettings{}
}

func (ct *FriendSettings) GetFriendSettings(c *gin.Context) {
	var req dto.FriendSettingsRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	friendSettings, err := service.NewFriendSettingsService(c).GetFriendSettings(req.ContactID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	if friendSettings == nil {
		resp.ToResponse(model.FriendSettings{})
		return
	}
	resp.ToResponse(friendSettings)
}

func (ct *FriendSettings) SaveFriendSettings(c *gin.Context) {
	var req model.FriendSettings
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	err := service.NewFriendSettingsService(c).SaveFriendSettings(&req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
