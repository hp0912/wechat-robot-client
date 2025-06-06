package common_cron

import (
	"context"
	"log"
	"net/http"
	"time"
	"wechat-robot-client/pkg/good_morning"
	"wechat-robot-client/service"
	"wechat-robot-client/vars"

	"github.com/go-resty/resty/v2"
)

type GoodMorningCron struct {
	CronManager *CronManager
}

func NewGoodMorningCron(cronManager *CronManager) vars.CommonCronInstance {
	return &GoodMorningCron{
		CronManager: cronManager,
	}
}

func (cron *GoodMorningCron) IsActive() bool {
	if cron.CronManager.globalSettings.MorningEnabled != nil && *cron.CronManager.globalSettings.MorningEnabled {
		return true
	}
	return false
}

func (cron *GoodMorningCron) Cron() error {
	// 获取当前时间
	now := time.Now()
	// 获取年、月、日
	year := now.Year()
	month := now.Month()
	day := now.Day()
	// 获取星期
	weekday := now.Weekday()
	// 定义中文星期数组
	weekdays := [...]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}

	chatRoomSettings, err := service.NewChatRoomSettingsService(context.Background()).GetAllEnableGoodMorning()
	if err != nil {
		log.Printf("获取群聊设置失败: %v", err)
		return err
	}

	// 每日一言
	dailyWords := "早上好，今天接口挂了，没有早安语。"
	resp, err := resty.New().R().
		Post("https://api.pearktrue.cn/api/hitokoto/")
	if err != nil || resp.StatusCode() != http.StatusOK {
		log.Printf("获取随机一言失败: %v", err)
	} else {
		respText := resp.String()
		if respText != "" {
			dailyWords = respText
		}
	}

	crService := service.NewChatRoomService(context.Background())
	msgService := service.NewMessageService(context.Background())
	for _, setting := range chatRoomSettings {
		summary, err := crService.GetChatRoomSummary(setting.ChatRoomID)
		if err != nil {
			log.Printf("统计群[%s]信息失败: %v", setting.ChatRoomID, err)
			continue
		}

		summary.Year = year
		summary.Month = int(month)
		summary.Date = day
		summary.Week = weekdays[weekday]

		image, err := good_morning.Draw(dailyWords, summary)
		if err != nil {
			log.Printf("绘制群[%s]早安图片失败: %v", setting.ChatRoomID, err)
			continue
		}

		err = msgService.MsgUploadImg(setting.ChatRoomID, image)
		if err != nil {
			log.Printf("群[%s]早安图片发送失败: %v", setting.ChatRoomID, err)
			continue
		}
		log.Printf("群[%s]早安图片发送成功", setting.ChatRoomID)
	}

	return nil
}

func (cron *GoodMorningCron) Register() {
	if !cron.IsActive() {
		log.Println("每日早安任务未启用")
		return
	}
	err := cron.CronManager.AddJob(vars.MorningCron, cron.CronManager.globalSettings.MorningCron, func() {
		log.Println("开始每日早安任务")
		if err := cron.Cron(); err != nil {
			log.Printf("每日早安任务执行失败: %v", err)
		} else {
			log.Println("每日早安任务执行完成")
		}
	})
	if err != nil {
		log.Printf("每日早安任务注册失败: %v", err)
		return
	}
	log.Println("每日早安任务初始化成功")
}
