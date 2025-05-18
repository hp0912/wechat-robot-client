package controller

import (
	"errors"
	"wechat-robot-client/model"
	"wechat-robot-client/pkg/appx"
	"wechat-robot-client/pkg/utils"
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
	if req.ChatRoomRankingEnabled != nil && *req.ChatRoomRankingEnabled {
		if req.ChatRoomRankingDailyCron == "" {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
		if req.ChatRoomRankingDailyCron != "" {
			if !utils.IsDailyAtHourMinute(req.ChatRoomRankingDailyCron) {
				resp.ToErrorResponse(errors.New("参数错误"))
				return
			}
		}
		if req.ChatRoomRankingWeeklyCron != nil && *req.ChatRoomRankingWeeklyCron != "" {
			if !utils.IsWeeklyMondayAtHourMinute(*req.ChatRoomRankingWeeklyCron) {
				resp.ToErrorResponse(errors.New("参数错误"))
				return
			}
		}
		if req.ChatRoomRankingMonthCron != nil && *req.ChatRoomRankingMonthCron != "" {
			if !utils.IsMonthly1stAtHourMinute(*req.ChatRoomRankingMonthCron) {
				resp.ToErrorResponse(errors.New("参数错误"))
				return
			}
		}
	}
	if req.ChatRoomSummaryEnabled != nil && *req.ChatRoomSummaryEnabled {
		if req.ChatRoomSummaryModel == "" || req.ChatRoomSummaryCron == "" {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
		if !utils.IsDailyAtHourMinute(req.ChatRoomSummaryCron) {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
	}
	if req.NewsEnabled != nil && *req.NewsEnabled {
		if req.NewsType == "" || req.NewsCron == "" {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
		if !utils.IsDailyAtHourMinute(req.NewsCron) {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
	}
	if req.MorningEnabled != nil && *req.MorningEnabled {
		if req.MorningCron == "" {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
		if !utils.IsDailyAtHourMinute(req.MorningCron) {
			resp.ToErrorResponse(errors.New("参数错误"))
			return
		}
	}
	service.NewGlobalSettingsService(c).SaveGlobalSettings(&req)
	resp.ToResponse(nil)
}
