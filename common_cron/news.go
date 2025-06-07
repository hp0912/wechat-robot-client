package common_cron

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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

func (cron *NewsCron) Cron() error {
	globalSettings, err := service.NewGlobalSettingsService(context.Background()).GetGlobalSettings()
	if err != nil {
		log.Printf("获取全局设置失败: %v", err)
		return err
	}
	settings, err := service.NewChatRoomSettingsService(context.Background()).GetAllEnableNews()
	if err != nil {
		log.Printf("获取群聊设置失败: %v", err)
		return err
	}

	var newsResp NewsResponse
	_, err = resty.New().R().
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
	resp, err := resty.New().R().SetDoNotParseResponse(true).Get(newsImage)
	if err != nil {
		return err
	}
	defer resp.RawBody().Close()
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "news_image_*")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name()) // 清理临时文件
	// 将图片数据写入临时文件
	_, err = io.Copy(tempFile, resp.RawBody())
	if err != nil {
		return err
	}

	for _, setting := range settings {
		newsType := globalSettings.NewsType
		if setting.NewsType != nil && *setting.NewsType != "" {
			newsType = *setting.NewsType
		}
		if newsType == "text" {
			err := msgService.SendTextMessage(setting.ChatRoomID, newsText)
			if err != nil {
				log.Printf("[每日早报] 发送文本消息失败: %v", err)
			}
		} else {
			// 重置文件指针到开始位置
			_, err := tempFile.Seek(0, 0)
			if err != nil {
				log.Printf("[每日早报] 重置文件指针失败: %v", err)
				continue
			}
			err = msgService.MsgUploadImg(setting.ChatRoomID, tempFile)
			if err != nil {
				log.Printf("[每日早报] 发送图片消息失败: %v", err)
			}
		}
	}

	return nil
}

func (cron *NewsCron) Register() {
	if !cron.IsActive() {
		log.Println("每日早报任务未启用")
		return
	}
	err := cron.CronManager.AddJob(vars.NewsCron, cron.CronManager.globalSettings.NewsCron, func() {
		log.Println("开始执行每日早报任务")
		if err := cron.Cron(); err != nil {
			log.Printf("执行每日早报任务失败: %v", err)
		} else {
			log.Println("每日早报任务执行完成")
		}
	})
	if err != nil {
		log.Printf("每日早报任务注册失败: %v", err)
		return
	}
	log.Println("每日早报任务初始化成功")
}
