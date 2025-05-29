package common_cron

import (
	"context"
	"fmt"
	"log"
	"strings"
	"wechat-robot-client/dto"
	"wechat-robot-client/model"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
)

type NewsCron struct {
	CronManager    *CronManager
	GlobalSettings *model.GlobalSettings
}

type NewsResponse struct {
	Code    string   `json:"code"`
	Msg     string   `json:"msg"`
	Date    string   `json:"date"`
	HeadImg string   `json:"head_image"`
	Image   string   `json:"image"`
	News    []string `json:"news"`
	Weiyu   string   `json:"weiyu"`
}

func NewNewsCron(cronManager *CronManager, globalSettings *model.GlobalSettings) *NewsCron {
	return &NewsCron{
		CronManager:    cronManager,
		GlobalSettings: globalSettings,
	}
}

func (cron *NewsCron) IsActive() bool {
	if cron.GlobalSettings.NewsEnabled != nil && *cron.GlobalSettings.NewsEnabled {
		return true
	}
	return false
}

func (cron *NewsCron) Start() {
	if !cron.IsActive() {
		return
	}
	cron.CronManager.AddJob(vars.FriendSyncCron, cron.GlobalSettings.FriendSyncCron, func(params ...any) error {
		log.Println("开始执行每日早报任务")

		settings := service.NewChatRoomSettingsService(context.Background()).GetAllEnableNews(vars.RobotRuntime.WxID)

		var newsResp NewsResponse
		_, err := resty.New().R().
			SetHeader("Content-Type", "application/json;chartset=utf-8").
			SetQueryParam("type", "json").
			SetResult(&newsResp).
			Get("https://api.suxun.site/api/sixs")
		if err != nil {
			return err
		}
		if newsResp.Code != "200" {
			return fmt.Errorf("获取每日早报失败: %s", newsResp.Msg)
		}

		msgService := service.NewMessageService(context.Background())
		newsText := strings.Join(newsResp.News, "\n")

		newsImage := newsResp.Image
		resp, err := resty.New().R().Get(newsImage)
		if err != nil {
			return err
		}
		defer resp.RawBody().Close()

		for _, setting := range settings {
			if setting.NewsType == "text" {
				msgService.SendTextMessage(dto.SendTextMessageRequest{
					SendMessageCommonRequest: dto.SendMessageCommonRequest{
						ToWxid: setting.ChatRoomID,
					},
					Content: newsText,
				})
			} else {
				msgService.MsgUploadImg(setting.ChatRoomID, resp.RawBody())
			}
		}

		return nil
	})
	log.Println("每日早报任务初始化成功")
}
