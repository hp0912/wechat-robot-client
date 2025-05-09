package task

import (
	"context"
	"log"
	"time"
	"wechat-robot-client/service"

	"github.com/go-co-op/gocron"
)

func InitTasks() {
	// 定时任务发送消息
	s := gocron.NewScheduler(time.Local)
	// 更新好友列表
	_, _ = s.Cron("0 */1 * * *").Do(service.NewContactService(context.Background()).SyncContact(true))
	// 开启定时任务
	s.StartAsync()
	log.Println("定时任务初始化成功")
}
