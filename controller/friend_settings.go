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
	friendSettings := service.NewFriendSettingsService(c).GetFriendSettings(req.ContactID)
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
	if req.ChatAIEnabled != nil && *req.ChatAIEnabled {
		if req.ChatAPIKey == "" || req.ChatBaseURL == "" || req.ChatModel == "" || req.ChatPrompt == "" {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
	}
	if req.ImageAIEnabled != nil && *req.ImageAIEnabled {
		if req.ImageModel == "" || req.ImageAISettings == nil {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
	}
	service.NewFriendSettingsService(c).SaveFriendSettings(&req)
	resp.ToResponse(nil)
}
