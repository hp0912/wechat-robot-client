package common_cron

import (
	"context"
	"fmt"
	"log"
	"strings"
	"wechat-robot-client/dto"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
)

type NewsCron struct {
	CronManager *CronManager
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

func NewNewsCron(cronManager *CronManager) vars.CommonCronInstance {
	return &NewsCron{
		CronManager: cronManager,
	}
}

func (cron *NewsCron) IsActive() bool {
	if cron.CronManager.globalSettings.NewsEnabled != nil && *cron.CronManager.globalSettings.NewsEnabled {
		return true
	}
	return false
}

func (cron *NewsCron) Register() {
	if !cron.IsActive() {
		log.Println("每日早报任务未启用")
		return
	}
	err := cron.CronManager.AddJob(vars.NewsCron, cron.CronManager.globalSettings.NewsCron, func() error {
		log.Println("开始执行每日早报任务")

		settings := service.NewChatRoomSettingsService(context.Background()).GetAllEnableNews()

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
				err := msgService.SendTextMessage(dto.SendTextMessageRequest{
					SendMessageCommonRequest: dto.SendMessageCommonRequest{
						ToWxid: setting.ChatRoomID,
					},
					Content: newsText,
				})
				if err != nil {
					log.Printf("[每日早报] 发送文本消息失败: %v", err)
				}
			} else {
				err := msgService.MsgUploadImg(setting.ChatRoomID, resp.RawBody())
				if err != nil {
					log.Printf("[每日早报] 发送图片消息失败: %v", err)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Printf("每日早报任务注册失败: %v", err)
		return
	}
	log.Println("每日早报任务初始化成功")
}
