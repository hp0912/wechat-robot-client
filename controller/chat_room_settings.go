package controller

import (
	"errors"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/service"

	"github.com/gin-gonic/gin"
)

type ChatRoomSettings struct {
}

func NewChatRoomSettingsController() *ChatRoomSettings {
	return &ChatRoomSettings{}
}

func (ct *ChatRoomSettings) GetChatRoomSettings(c *gin.Context) {
	var req dto.ChatRoomSettingsRequest
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	chatRoomSettings, err := service.NewChatRoomSettingsService(c).GetChatRoomSettings(req.ChatRoomID)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	if chatRoomSettings == nil {
		resp.ToResponse(model.ChatRoomSettings{})
		return
	}
	resp.ToResponse(chatRoomSettings)
}

func (ct *ChatRoomSettings) SaveChatRoomSettings(c *gin.Context) {
	var req model.ChatRoomSettings
	resp := appx.NewResponse(c)
	if ok, err := appx.BindAndValid(c, &req); !ok || err != nil {
		resp.ToErrorResponse(errors.New("参数错误"))
		return
	}
	if req.WelcomeEnabled != nil && *req.WelcomeEnabled {
		if req.WelcomeType == "" {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
		if req.WelcomeType == model.WelcomeTypeText {
			if req.WelcomeText == "" {
				resp.ToErrorResponse(errors.New("参数错误"))
				return
			}
		}
		if req.WelcomeType == model.WelcomeTypeEmoji {
			if req.WelcomeEmojiMD5 == "" || req.WelcomeEmojiLen == 0 {
				resp.ToErrorResponse(errors.New("参数错误"))
				return
			}
		}
		if req.WelcomeType == model.WelcomeTypeImage {
			if req.WelcomeImageURL == "" {
				resp.ToErrorResponse(errors.New("参数错误"))
				return
			}
		}
		if req.WelcomeType == model.WelcomeTypeURL {
			if req.WelcomeText == "" || req.WelcomeURL == "" {
				resp.ToErrorResponse(errors.New("参数错误"))
				return
			}
		}
	}
	err := service.NewChatRoomSettingsService(c).SaveChatRoomSettings(&req)
	if err != nil {
		resp.ToErrorResponse(err)
		return
	}
	resp.ToResponse(nil)
}
