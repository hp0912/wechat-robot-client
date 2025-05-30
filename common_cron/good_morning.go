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

func (cron *GoodMorningCron) Register() {
	if !cron.IsActive() {
		log.Println("每日早安任务未启用")
		return
	}
	cron.CronManager.AddJob(vars.FriendSyncCron, cron.CronManager.globalSettings.FriendSyncCron, func(params ...any) error {
		log.Println("开始每日早安任务")

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

		chatRoomSettings := service.NewChatRoomSettingsService(context.Background()).GetAllEnableGoodMorning(vars.RobotRuntime.WxID)

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
			time.Sleep(1 * time.Second) // 避免发送过快
		}

		return nil
	})
	log.Println("每日早安任务初始化成功")
}
